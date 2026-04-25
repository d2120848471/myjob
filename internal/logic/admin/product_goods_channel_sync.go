package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/gogf/gf/v2/frame/g"
)

const defaultProductGoodsChannelSyncLimit = 200

var errProductInfoProviderUnsupported = errors.New("供应商不支持商品详情同步")

// ProductGoodsChannelSyncOptions 控制商品渠道同步范围。
type ProductGoodsChannelSyncOptions struct {
	GoodsID int64
	Limit   int
}

// ProductGoodsChannelSyncResult 汇总一次商品渠道同步的处理结果。
type ProductGoodsChannelSyncResult struct {
	Scanned int
	Updated int
	Skipped int
	Failed  int
}

type productGoodsChannelSyncCandidate struct {
	BindingID              int64  `db:"binding_id"`
	GoodsID                int64  `db:"goods_id"`
	GoodsHasTax            int    `db:"goods_has_tax"`
	PlatformAccountID      int64  `db:"platform_account_id"`
	ChannelHasTax          int    `db:"channel_has_tax"`
	ProviderCode           string `db:"provider_code"`
	Domain                 string `db:"domain"`
	BackupDomain           string `db:"backup_domain"`
	TokenID                string `db:"token_id"`
	SecretKey              string `db:"secret_key"`
	ExtraConfig            string `db:"extra_config"`
	SupplierGoodsNo        string `db:"supplier_goods_no"`
	SyncCostPriceEnabled   int    `db:"sync_cost_price_enabled"`
	SyncGoodsNameEnabled   int    `db:"sync_goods_name_enabled"`
	CurrentSupplierName    string `db:"supplier_goods_name"`
	CurrentSourceCostPrice string `db:"source_cost_price"`
}

// SyncChannelBindingsOnce 同步开启了商品维度开关的渠道绑定商品名和进货价。
func (l *ProductGoodsLogic) SyncChannelBindingsOnce(ctx context.Context, opts ProductGoodsChannelSyncOptions) (ProductGoodsChannelSyncResult, error) {
	limit := normalizeProductGoodsChannelSyncLimit(opts.Limit)
	var result ProductGoodsChannelSyncResult
	cache := make(map[string]supplierprovider.ProductInfoResult)
	afterBindingID := int64(0)
	for {
		candidates, err := l.loadProductGoodsChannelSyncCandidates(ctx, opts, afterBindingID, limit)
		if err != nil {
			return ProductGoodsChannelSyncResult{}, err
		}
		if len(candidates) == 0 {
			break
		}
		result.Scanned += len(candidates)
		for _, candidate := range candidates {
			afterBindingID = candidate.BindingID
			info, fetchErr := l.fetchProductGoodsChannelProductInfo(ctx, candidate, cache)
			if fetchErr != nil {
				if errors.Is(fetchErr, errProductInfoProviderUnsupported) {
					result.Skipped++
					continue
				}
				g.Log().Warningf(ctx, "商品渠道信息同步请求失败：binding=%d provider=%s goods_no=%s error=%v", candidate.BindingID, candidate.ProviderCode, candidate.SupplierGoodsNo, fetchErr)
				result.Failed++
				continue
			}
			updated, applyErr := l.applyProductGoodsChannelProductInfo(ctx, candidate, info)
			if applyErr != nil {
				g.Log().Warningf(ctx, "商品渠道信息同步保存失败：binding=%d provider=%s goods_no=%s error=%v", candidate.BindingID, candidate.ProviderCode, candidate.SupplierGoodsNo, applyErr)
				result.Failed++
				continue
			}
			if updated {
				result.Updated++
			} else {
				result.Skipped++
			}
		}
		if len(candidates) < limit {
			break
		}
	}
	return result, nil
}

func normalizeProductGoodsChannelSyncLimit(limit int) int {
	if limit <= 0 || limit > defaultProductGoodsChannelSyncLimit {
		return defaultProductGoodsChannelSyncLimit
	}
	return limit
}

func (l *ProductGoodsLogic) loadProductGoodsChannelSyncCandidates(ctx context.Context, opts ProductGoodsChannelSyncOptions, afterBindingID int64, limit int) ([]productGoodsChannelSyncCandidate, error) {
	rows := make([]productGoodsChannelSyncCandidate, 0)
	err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id AS binding_id,
    b.goods_id,
    g.has_tax AS goods_has_tax,
    b.platform_account_id,
    a.has_tax AS channel_has_tax,
    a.provider_code,
    a.domain,
    a.backup_domain,
    a.token_id,
    a.secret_key,
    a.extra_config,
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.source_cost_price,
    c.sync_cost_price_enabled,
    c.sync_goods_name_enabled
FROM product_goods_channel_binding b
JOIN product_goods g ON g.id = b.goods_id
JOIN product_goods_channel_config c ON c.goods_id = b.goods_id
JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.is_deleted = 0
  AND g.is_deleted = 0
  AND g.status = 1
  AND g.supply_type = 'channel'
  AND a.is_deleted = 0
  AND a.status = 1
  AND b.supplier_goods_no <> ''
  AND (c.sync_cost_price_enabled = 1 OR c.sync_goods_name_enabled = 1)
  AND (? = 0 OR b.goods_id = ?)
  AND (? = 0 OR b.id > ?)
ORDER BY b.id ASC
LIMIT ?
`, opts.GoodsID, opts.GoodsID, afterBindingID, afterBindingID, limit)
	return rows, err
}

func (l *ProductGoodsLogic) fetchProductGoodsChannelProductInfo(ctx context.Context, candidate productGoodsChannelSyncCandidate, cache map[string]supplierprovider.ProductInfoResult) (supplierprovider.ProductInfoResult, error) {
	provider, ok := supplierprovider.LookupProductInfo(candidate.ProviderCode)
	if !ok {
		return supplierprovider.ProductInfoResult{}, errProductInfoProviderUnsupported
	}
	cacheKey := fmt.Sprintf("%d:%s", candidate.PlatformAccountID, candidate.SupplierGoodsNo)
	if cached, exists := cache[cacheKey]; exists {
		return cached, nil
	}
	extraConfig, err := parseExtraConfig(candidate.ExtraConfig)
	if err != nil {
		return supplierprovider.ProductInfoResult{}, err
	}
	account := supplierprovider.AccountConfig{
		ProviderCode: candidate.ProviderCode,
		Domain:       candidate.Domain,
		BackupDomain: candidate.BackupDomain,
		TokenID:      candidate.TokenID,
		SecretKey:    candidate.SecretKey,
		ExtraConfig:  extraConfig,
	}
	client := l.httpClientForProductInfoProvider(candidate.ProviderCode)
	var lastErr error
	for _, baseURL := range provider.CandidateBaseURLs(account) {
		request, buildErr := provider.BuildProductInfoRequest(ctx, account, l.core.Now(), baseURL, supplierprovider.ProductInfoInput{SupplierGoodsNo: candidate.SupplierGoodsNo})
		if buildErr != nil {
			return supplierprovider.ProductInfoResult{}, buildErr
		}
		response, requestErr := client.Do(request)
		if requestErr != nil {
			lastErr = requestErr
			continue
		}
		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if readErr != nil {
			return supplierprovider.ProductInfoResult{}, readErr
		}
		info, parseErr := provider.ParseProductInfoResponse(response.StatusCode, body)
		if parseErr != nil {
			lastErr = parseErr
			continue
		}
		cache[cacheKey] = info
		return info, nil
	}
	if lastErr != nil {
		return supplierprovider.ProductInfoResult{}, lastErr
	}
	return supplierprovider.ProductInfoResult{}, errors.New("供应商商品详情候选地址为空")
}

func (l *ProductGoodsLogic) applyProductGoodsChannelProductInfo(ctx context.Context, candidate productGoodsChannelSyncCandidate, info supplierprovider.ProductInfoResult) (bool, error) {
	if info.SupplierGoodsNo != "" && strings.TrimSpace(info.SupplierGoodsNo) != candidate.SupplierGoodsNo {
		g.Log().Warningf(ctx, "商品渠道信息同步跳过：binding=%d goods_no=%s upstream_goods_no=%s error=%s", candidate.BindingID, candidate.SupplierGoodsNo, info.SupplierGoodsNo, "上游商品编号不一致")
		return false, nil
	}
	setParts := make([]string, 0, 6)
	args := make([]any, 0, 8)
	requiredSwitches := make([]string, 0, 2)
	if candidate.SyncGoodsNameEnabled == 1 {
		name := strings.TrimSpace(info.GoodsName)
		if name != "" {
			setParts = append(setParts, "supplier_goods_name = ?")
			args = append(args, name)
			requiredSwitches = append(requiredSwitches, "c.sync_goods_name_enabled = 1")
		}
	}
	if candidate.SyncCostPriceEnabled == 1 {
		if !info.GoodsPriceValid {
			g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%s", candidate.BindingID, candidate.SupplierGoodsNo, "上游价格无效")
		} else {
			financeTaxConfig, err := l.loadProductGoodsChannelFinanceTaxConfig(ctx, candidate.GoodsHasTax, candidate.ChannelHasTax)
			if err != nil {
				g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%v", candidate.BindingID, candidate.SupplierGoodsNo, err)
			} else {
				// 价格同步只改上游进货价和税价快照，自动改价利润字段保持用户配置不变。
				snapshot, snapshotErr := computeChannelCostSnapshot(info.GoodsPrice.StringFixed(4), candidate.GoodsHasTax, candidate.ChannelHasTax, financeTaxConfig)
				if snapshotErr == nil {
					setParts = append(setParts,
						"source_cost_price = ?",
						"cost_price = ?",
						"tax_adjust_direction = ?",
						"tax_adjust_rate = ?",
						"tax_adjust_amount = ?",
					)
					args = append(args, snapshot.SourceCostPrice, snapshot.CostPrice, snapshot.TaxAdjustDirection, snapshot.TaxAdjustRate, snapshot.TaxAdjustAmount)
					requiredSwitches = append(requiredSwitches, "c.sync_cost_price_enabled = 1")
				} else {
					g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%v", candidate.BindingID, candidate.SupplierGoodsNo, snapshotErr)
				}
			}
		}
	}
	if len(setParts) == 0 {
		return false, nil
	}
	setParts = append(setParts, "updated_at = ?")
	args = append(args, l.core.Now(), candidate.BindingID, candidate.PlatformAccountID, candidate.SupplierGoodsNo, candidate.PlatformAccountID)
	result, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET `+strings.Join(setParts, ", ")+`
WHERE id = ?
  AND is_deleted = 0
  AND platform_account_id = ?
  AND supplier_goods_no = ?
  AND EXISTS (
      SELECT 1
      FROM product_goods g
      JOIN product_goods_channel_config c ON c.goods_id = g.id
      JOIN supplier_platform_account a ON a.id = ?
      WHERE g.id = product_goods_channel_binding.goods_id
        AND g.is_deleted = 0
        AND g.status = 1
        AND g.supply_type = 'channel'
        AND a.is_deleted = 0
        AND a.status = 1
        AND `+strings.Join(requiredSwitches, " AND ")+`
  )
`, args...)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (l *ProductGoodsLogic) httpClientForProductInfoProvider(providerCode string) *http.Client {
	baseClient := l.httpClient
	if baseClient == nil {
		baseClient = http.DefaultClient
	}
	if providerCode != "kakayun" {
		return baseClient
	}
	if baseClient.Transport != nil {
		if _, ok := baseClient.Transport.(*http.Transport); !ok {
			// 测试或调用方可能注入自定义 RoundTripper，此时不能强行替换 transport。
			return baseClient
		}
	}
	client := *baseClient
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := baseClient.Transport.(*http.Transport); ok && baseTransport != nil {
		transport = baseTransport.Clone()
	}
	// 卡卡云 dock 接口对压缩响应兼容性不稳定，同步请求显式关闭压缩避免 EOF 误判。
	transport.DisableCompression = true
	client.Transport = transport
	return &client
}
