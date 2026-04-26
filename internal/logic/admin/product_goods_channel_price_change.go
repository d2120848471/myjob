package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"
)

const (
	productGoodsChannelPriceChangeSourceMonitor = "monitor"
	productGoodsChannelPriceChangeSourcePush    = "push"
)

type productGoodsChannelPriceChangeApplyResult struct {
	Updated      bool
	PriceChanged bool
}

func (l *ProductGoodsLogic) applyProductGoodsChannelPriceChange(ctx context.Context, candidate productGoodsChannelSyncCandidate, info supplierprovider.ProductInfoResult, source string) (productGoodsChannelPriceChangeApplyResult, error) {
	if info.SupplierGoodsNo != "" && strings.TrimSpace(info.SupplierGoodsNo) != candidate.SupplierGoodsNo {
		g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s upstream_goods_no=%s error=%s", candidate.BindingID, candidate.SupplierGoodsNo, info.SupplierGoodsNo, "上游商品编号不一致")
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}
	if candidate.SyncCostPriceEnabled != 1 {
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}
	if !info.GoodsPriceValid {
		g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%s", candidate.BindingID, candidate.SupplierGoodsNo, "上游价格无效")
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}

	financeTaxConfig, err := l.loadProductGoodsChannelFinanceTaxConfig(ctx, candidate.GoodsHasTax, candidate.ChannelHasTax)
	if err != nil {
		g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%v", candidate.BindingID, candidate.SupplierGoodsNo, err)
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}
	// 价格同步只改上游进货价和税价快照；利润配置保持用户设置，是否自动改价由读取端按配置实时计算。
	snapshot, err := computeChannelCostSnapshot(info.GoodsPrice.StringFixed(4), candidate.GoodsHasTax, candidate.ChannelHasTax, financeTaxConfig)
	if err != nil {
		g.Log().Warningf(ctx, "商品渠道进价同步跳过：binding=%d goods_no=%s error=%v", candidate.BindingID, candidate.SupplierGoodsNo, err)
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}

	oldSourceCostPrice := priceChangeMoneyOrZero(candidate.CurrentSourceCostPrice)
	oldCostPrice := priceChangeMoneyOrZero(candidate.CurrentCostPrice)
	newSourceCostPrice := priceChangeMoneyOrZero(snapshot.SourceCostPrice)
	newCostPrice := priceChangeMoneyOrZero(snapshot.CostPrice)
	if oldSourceCostPrice == newSourceCostPrice && oldCostPrice == newCostPrice {
		return productGoodsChannelPriceChangeApplyResult{}, nil
	}

	oldEffectiveSellPrice := l.computePriceChangeEffectiveSellPrice(ctx, candidate, oldCostPrice)
	newEffectiveSellPrice := l.computePriceChangeEffectiveSellPrice(ctx, candidate, newCostPrice)
	changeAmount := priceChangeAmount(oldEffectiveSellPrice, newEffectiveSellPrice)
	description := buildProductGoodsChannelPriceChangeDescription(source, candidate, oldSourceCostPrice, newSourceCostPrice, oldCostPrice, newCostPrice, oldEffectiveSellPrice, newEffectiveSellPrice)
	supplierGoodsName := strings.TrimSpace(info.GoodsName)
	if supplierGoodsName == "" {
		supplierGoodsName = candidate.CurrentSupplierName
	}

	result := productGoodsChannelPriceChangeApplyResult{PriceChanged: true}
	now := l.core.Now()
	err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		updateResult, updateErr := tx.Exec(`
UPDATE product_goods_channel_binding
SET source_cost_price = ?,
    cost_price = ?,
    tax_adjust_direction = ?,
    tax_adjust_rate = ?,
    tax_adjust_amount = ?,
    updated_at = ?
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
        AND c.sync_cost_price_enabled = 1
  )
`, newSourceCostPrice, newCostPrice, snapshot.TaxAdjustDirection, snapshot.TaxAdjustRate, snapshot.TaxAdjustAmount, now, candidate.BindingID, candidate.PlatformAccountID, candidate.SupplierGoodsNo, candidate.PlatformAccountID)
		if updateErr != nil {
			return updateErr
		}
		affected, updateErr := updateResult.RowsAffected()
		if updateErr != nil {
			return updateErr
		}
		if affected == 0 {
			result.PriceChanged = false
			return nil
		}
		result.Updated = true
		return nil
	})
	if err != nil {
		return productGoodsChannelPriceChangeApplyResult{}, err
	}
	if !result.Updated {
		return result, nil
	}

	// 改价记录是审计辅助信息，插入失败不能回滚已经完成的渠道进价更新。
	if _, insertErr := l.core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_price_change_log (
    source, provider_code, platform_account_id, platform_account_name, binding_id,
    goods_id, goods_code, goods_name, goods_icon, supplier_goods_no, supplier_goods_name,
    old_source_cost_price, new_source_cost_price, old_cost_price, new_cost_price,
    old_effective_sell_price, new_effective_sell_price, change_amount,
    description, raw_payload, changed_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, source, candidate.ProviderCode, candidate.PlatformAccountID, candidate.PlatformAccountName, candidate.BindingID,
		candidate.GoodsID, candidate.GoodsCode, candidate.GoodsName, candidate.GoodsIcon, candidate.SupplierGoodsNo, supplierGoodsName,
		oldSourceCostPrice, newSourceCostPrice, oldCostPrice, newCostPrice,
		oldEffectiveSellPrice, newEffectiveSellPrice, changeAmount,
		description, info.Raw, now, now); insertErr != nil {
		g.Log().Warningf(ctx, "商品渠道改价记录写入失败：binding=%d provider=%s goods_no=%s error=%v", candidate.BindingID, candidate.ProviderCode, candidate.SupplierGoodsNo, insertErr)
	}
	return result, nil
}

// HandleSupplierProductChangeCallback 处理第三方商品变动推送并复用渠道进价同步规则。
func (l *ProductGoodsLogic) HandleSupplierProductChangeCallback(ctx context.Context, req *adminapi.SupplierProductChangeCallbackReq, body []byte) error {
	providerCode := strings.TrimSpace(strings.ToLower(req.ProviderCode))
	account, err := l.getActiveSupplierPlatformAccount(ctx, req.PlatformAccountID)
	if err != nil {
		return apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	if account.Status != consts.StatusEnabled {
		return apiErr(consts.CodeBadRequest, "渠道账号已关闭")
	}
	if !strings.EqualFold(account.ProviderCode, providerCode) {
		return apiErr(consts.CodeBadRequest, "渠道账号与回调平台不匹配")
	}
	provider, ok := supplierprovider.LookupProductChangePush(providerCode)
	if !ok {
		return apiErr(consts.CodeBadRequest, "供应商不支持商品变动推送")
	}
	extraConfig, err := parseExtraConfig(account.ExtraConfig)
	if err != nil {
		return err
	}
	push, err := provider.ParseProductChangePush(supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  extraConfig,
	}, l.core.Now(), body)
	if err != nil {
		return err
	}

	candidates, err := l.loadProductGoodsChannelPushCandidates(ctx, account.ID, push.SupplierGoodsNo)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		_, err = l.applyProductGoodsChannelPriceChange(ctx, candidate, supplierprovider.ProductInfoResult{
			SupplierGoodsNo: push.SupplierGoodsNo,
			GoodsName:       push.GoodsName,
			GoodsPrice:      push.GoodsPrice,
			GoodsPriceValid: push.GoodsPriceValid,
			Raw:             push.Raw,
		}, productGoodsChannelPriceChangeSourcePush)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *ProductGoodsLogic) loadProductGoodsChannelPushCandidates(ctx context.Context, platformAccountID int64, supplierGoodsNo string) ([]productGoodsChannelSyncCandidate, error) {
	rows := make([]productGoodsChannelSyncCandidate, 0)
	err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id AS binding_id,
    b.goods_id,
    g.goods_code,
    g.name AS goods_name,
    COALESCE(pb.icon, '') AS goods_icon,
    g.has_tax AS goods_has_tax,
    g.default_sell_price,
    b.platform_account_id,
    a.name AS platform_account_name,
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
    b.cost_price,
    b.tax_adjust_direction,
    b.tax_adjust_rate,
    b.tax_adjust_amount,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    c.sync_cost_price_enabled,
    c.sync_goods_name_enabled
FROM product_goods_channel_binding b
JOIN product_goods g ON g.id = b.goods_id
LEFT JOIN product_brand pb ON pb.id = g.brand_id
JOIN product_goods_channel_config c ON c.goods_id = b.goods_id
JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.is_deleted = 0
  AND g.is_deleted = 0
  AND g.status = 1
  AND g.supply_type = 'channel'
  AND a.is_deleted = 0
  AND a.status = 1
  AND b.platform_account_id = ?
  AND b.supplier_goods_no = ?
ORDER BY b.id ASC
`, platformAccountID, strings.TrimSpace(supplierGoodsNo))
	return rows, err
}

// ListProductGoodsChannelPriceChanges 分页查询监控或推送触发的商品渠道改价记录。
func (l *ProductGoodsLogic) ListProductGoodsChannelPriceChanges(ctx context.Context, req *adminapi.ProductGoodsChannelPriceChangeListReq) (*adminapi.ProductGoodsChannelPriceChangeListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 12)
	if source := strings.TrimSpace(req.Source); source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, source)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		conditions = append(conditions, "(goods_code LIKE ? OR goods_name LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	if supplierGoodsNo := strings.TrimSpace(req.SupplierGoodsNo); supplierGoodsNo != "" {
		conditions = append(conditions, "supplier_goods_no = ?")
		args = append(args, supplierGoodsNo)
	}
	if req.PlatformID > 0 {
		conditions = append(conditions, "platform_account_id = ?")
		args = append(args, req.PlatformID)
	}
	if startAt := strings.TrimSpace(req.StartAt); startAt != "" {
		conditions = append(conditions, "changed_at >= ?")
		args = append(args, startAt)
	}
	if endAt := strings.TrimSpace(req.EndAt); endAt != "" {
		conditions = append(conditions, "changed_at <= ?")
		args = append(args, endAt)
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_price_change_log WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "自动改价记录查询失败")
	}
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    id, source, provider_code, platform_account_id, platform_account_name, binding_id,
    goods_id, goods_code, goods_name, goods_icon, supplier_goods_no, supplier_goods_name,
    old_source_cost_price, new_source_cost_price, old_cost_price, new_cost_price,
    old_effective_sell_price, new_effective_sell_price, change_amount,
    description, raw_payload, changed_at
FROM product_goods_channel_price_change_log
WHERE `+whereClause+`
ORDER BY changed_at DESC, id DESC
LIMIT ? OFFSET ?
`, queryArgs...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "自动改价记录查询失败")
	}

	items := make([]adminapi.ProductGoodsChannelPriceChangeItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminapi.ProductGoodsChannelPriceChangeItem{
			ID:                    row["id"].Int64(),
			Source:                row["source"].String(),
			ProviderCode:          row["provider_code"].String(),
			PlatformAccountID:     row["platform_account_id"].Int64(),
			PlatformAccountName:   row["platform_account_name"].String(),
			BindingID:             row["binding_id"].Int64(),
			GoodsID:               row["goods_id"].Int64(),
			GoodsCode:             row["goods_code"].String(),
			GoodsName:             row["goods_name"].String(),
			GoodsIcon:             row["goods_icon"].String(),
			SupplierGoodsNo:       row["supplier_goods_no"].String(),
			SupplierGoodsName:     row["supplier_goods_name"].String(),
			OldSourceCostPrice:    productGoodsRecordMoney(row, "old_source_cost_price"),
			NewSourceCostPrice:    productGoodsRecordMoney(row, "new_source_cost_price"),
			OldCostPrice:          productGoodsRecordMoney(row, "old_cost_price"),
			NewCostPrice:          productGoodsRecordMoney(row, "new_cost_price"),
			OldEffectiveSellPrice: productGoodsRecordMoney(row, "old_effective_sell_price"),
			NewEffectiveSellPrice: productGoodsRecordMoney(row, "new_effective_sell_price"),
			ChangeAmount:          productGoodsRecordMoney(row, "change_amount"),
			Description:           row["description"].String(),
			RawPayload:            row["raw_payload"].String(),
			ChangedAt:             formatAppTime(parseRecordTime(row, "changed_at")),
		})
	}
	return &adminapi.ProductGoodsChannelPriceChangeListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

func (l *ProductGoodsLogic) computePriceChangeEffectiveSellPrice(ctx context.Context, candidate productGoodsChannelSyncCandidate, costPrice string) string {
	price, err := computeChannelEffectiveSellPrice(candidate.DefaultSellPrice, costPrice, candidate.IsAutoChange, candidate.AddType, candidate.DefaultPrice)
	if err != nil {
		g.Log().Warningf(ctx, "商品渠道利润后价格计算失败：binding=%d goods_no=%s error=%v", candidate.BindingID, candidate.SupplierGoodsNo, err)
		return "0.0000"
	}
	return priceChangeMoneyOrZero(price)
}

func priceChangeMoneyOrZero(value string) string {
	value = formatMoney(value)
	if strings.TrimSpace(value) == "" {
		return "0.0000"
	}
	return value
}

func priceChangeAmount(oldPrice, newPrice string) string {
	oldAmount, oldErr := decimal.NewFromString(oldPrice)
	newAmount, newErr := decimal.NewFromString(newPrice)
	if oldErr != nil || newErr != nil {
		return "0.0000"
	}
	return newAmount.Sub(oldAmount).StringFixed(4)
}

func buildProductGoodsChannelPriceChangeDescription(source string, candidate productGoodsChannelSyncCandidate, oldSource, newSource, oldCost, newCost, oldEffective, newEffective string) string {
	return fmt.Sprintf(
		"来源:%s；货源:%s；上游商品:%s；进价:%s -> %s；比较成本:%s -> %s；利润后价格:%s -> %s",
		productGoodsChannelPriceChangeSourceLabel(source),
		candidate.PlatformAccountName,
		candidate.SupplierGoodsNo,
		oldSource,
		newSource,
		oldCost,
		newCost,
		oldEffective,
		newEffective,
	)
}

func productGoodsChannelPriceChangeSourceLabel(source string) string {
	switch source {
	case productGoodsChannelPriceChangeSourcePush:
		return "推送"
	case productGoodsChannelPriceChangeSourceMonitor:
		return "监控"
	default:
		return strings.TrimSpace(source)
	}
}
