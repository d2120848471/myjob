package tradelogic

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/shopspring/decimal"
)

// HandleProviderPriceNotify 处理上游价格通知：验签、幂等写日志、定位 binding，并刷新成本与商品摘要。
func (l *TradeOrderLogic) HandleProviderPriceNotify(ctx context.Context, providerCode string, headers http.Header, body []byte) ([]byte, string, error) {
	providerCode = strings.TrimSpace(strings.ToLower(providerCode))
	if providerCode == "" {
		return nil, "", apiErr(consts.CodeBadRequest, "provider_code不能为空")
	}

	pnProvider, ok := supplierprovider.LookupPriceNotify(providerCode)
	if !ok {
		return nil, "", apiErr(consts.CodeBadRequest, "provider不支持价格通知")
	}

	account, err := l.loadPlatformAccountForPriceNotify(ctx, providerCode, body)
	if err != nil {
		return nil, "", err
	}
	accountCfg := supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		ExtraConfig:  decodeJSONMap(strings.TrimSpace(account.ExtraConfig)),
	}

	verifyErr := pnProvider.VerifyPriceNotifySignature(accountCfg, headers, body)
	verifyResult := "ok"
	if verifyErr != nil {
		verifyResult = "failed"
	}

	result, parseErr := pnProvider.ParsePriceNotifyPayload(accountCfg, headers, body)
	if parseErr != nil || result == nil {
		return nil, "", apiErr(consts.CodeBadRequest, "价格通知解析失败")
	}
	if strings.TrimSpace(result.IdempotencyKey) == "" {
		sum := sha256.Sum256(body)
		result.IdempotencyKey = fmt.Sprintf("%x", sum[:])
	}

	ackBody, contentType := defaultPriceNotifyAck(providerCode)

	now := l.core.Now()
	headersSnapshot, _ := snapshotHeaders(headers, account.TokenID, account.SecretKey)
	bodySnapshot := truncateSnapshot(sanitizeSnapshot(string(body), account.TokenID, account.SecretKey))

	sourceCostPriceNew := MoneyString(Round4(result.SourceCostPrice))
	processResult := "processed"
	if verifyErr != nil {
		processResult = "verify_failed"
	}

	inserted, err := l.core.DB().Exec(ctx, `
INSERT INTO provider_price_notify_log (
    provider_code, platform_account_id, idempotency_key, supplier_goods_no,
    request_headers, request_body, source_cost_price_new,
    verify_result, process_result,
    created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`,
		providerCode,
		account.ID,
		strings.TrimSpace(result.IdempotencyKey),
		strings.TrimSpace(result.SupplierGoodsNo),
		headersSnapshot,
		bodySnapshot,
		sourceCostPriceNew,
		verifyResult,
		processResult,
		now,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return ackBody, contentType, nil
		}
		return nil, "", apiErr(consts.CodeInternalError, "价格通知日志写入失败")
	}
	logID, _ := inserted.LastInsertId()

	if verifyErr != nil {
		return ackBody, contentType, apiErr(consts.CodeBadRequest, "价格通知验签失败")
	}

	bindingID, goodsID, bindingErr := l.locateBindingForPriceNotify(ctx, account.ID, result.SupplierGoodsNo)
	if bindingErr != nil {
		processResult := "binding_not_found"
		if strings.Contains(strings.ToLower(bindingErr.Error()), "duplicated") {
			processResult = "binding_duplicated"
		}
		_ = l.updatePriceNotifyLogProcessResult(ctx, logID, processResult)
		return ackBody, contentType, nil
	}

	goodsHasTax, err := l.loadGoodsHasTax(ctx, goodsID)
	if err != nil {
		_ = l.updatePriceNotifyLogProcessResult(ctx, logID, "goods_not_found")
		return ackBody, contentType, err
	}

	adjust, err := l.calcBindingCostPrice(ctx, goodsHasTax, account.HasTax, Round4(result.SourceCostPrice))
	if err != nil {
		processResult := "tax_rate_missing"
		if strings.Contains(err.Error(), "税点格式错误") {
			processResult = "tax_rate_invalid"
		}
		_ = l.updatePriceNotifyLogProcessResult(ctx, logID, processResult)
		return ackBody, contentType, err
	}

	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET supplier_goods_name = CASE WHEN ? != '' THEN ? ELSE supplier_goods_name END,
    source_cost_price = ?,
    cost_price = ?,
    tax_adjust_direction = ?,
    tax_adjust_rate = ?,
    tax_adjust_amount = ?,
    updated_at = ?
WHERE id = ? AND is_deleted = 0
`,
		strings.TrimSpace(result.SupplierGoodsName),
		strings.TrimSpace(result.SupplierGoodsName),
		adjust.SourceCostPrice,
		adjust.CostPrice,
		adjust.TaxAdjustDirection,
		adjust.TaxAdjustRate,
		adjust.TaxAdjustAmount,
		now,
		bindingID,
	); err != nil {
		_ = l.updatePriceNotifyLogProcessResult(ctx, logID, "binding_update_failed")
		return ackBody, contentType, nil
	}

	_ = l.refreshGoodsChannelSummary(ctx, goodsID)
	if strings.TrimSpace(result.SupplierGoodsName) != "" {
		cfg, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT sync_goods_name_enabled, primary_binding_id
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
		if err == nil && cfg != nil && len(cfg) > 0 && cfg["sync_goods_name_enabled"].Int() != 0 {
			if cfg["primary_binding_id"] != nil && !cfg["primary_binding_id"].IsNil() && cfg["primary_binding_id"].Int64() == bindingID {
				_, _ = l.core.DB().Exec(ctx, `
UPDATE product_goods
SET name = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, strings.TrimSpace(result.SupplierGoodsName), now, goodsID)
			}
		}
	}
	return ackBody, contentType, nil
}

func (l *TradeOrderLogic) updatePriceNotifyLogProcessResult(ctx context.Context, logID int64, processResult string) error {
	if logID <= 0 || strings.TrimSpace(processResult) == "" {
		return nil
	}
	_, err := l.core.DB().Exec(ctx, `UPDATE provider_price_notify_log SET process_result = ? WHERE id = ?`, strings.TrimSpace(processResult), logID)
	return err
}

func defaultPriceNotifyAck(providerCode string) ([]byte, string) {
	switch strings.TrimSpace(strings.ToLower(providerCode)) {
	case "xingquanyi":
		return []byte("ok"), "text/plain"
	default:
		return nil, "text/plain"
	}
}

func (l *TradeOrderLogic) loadPlatformAccountForPriceNotify(ctx context.Context, providerCode string, body []byte) (entity.SupplierPlatformAccount, error) {
	locator := extractCallbackAccountLocator(providerCode, body)
	query := `SELECT * FROM supplier_platform_account WHERE provider_code = ? AND is_deleted = 0 ORDER BY id ASC LIMIT 1`
	args := []any{providerCode}
	if strings.TrimSpace(locator) != "" {
		query = `SELECT * FROM supplier_platform_account WHERE provider_code = ? AND token_id = ? AND is_deleted = 0 ORDER BY id ASC LIMIT 1`
		args = []any{providerCode, strings.TrimSpace(locator)}
	}
	account := entity.SupplierPlatformAccount{}
	if err := l.core.DB().GetCore().GetScan(ctx, &account, query, args...); err != nil {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeInternalError, "读取渠道账号失败")
	}
	if account.ID <= 0 {
		return entity.SupplierPlatformAccount{}, apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	return account, nil
}

func (l *TradeOrderLogic) locateBindingForPriceNotify(ctx context.Context, platformAccountID int64, supplierGoodsNo string) (int64, int64, error) {
	supplierGoodsNo = strings.TrimSpace(supplierGoodsNo)
	if platformAccountID <= 0 || supplierGoodsNo == "" {
		return 0, 0, fmt.Errorf("invalid")
	}

	type row struct {
		ID      int64 `db:"id"`
		GoodsID int64 `db:"goods_id"`
	}
	rows := make([]row, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, goods_id
FROM product_goods_channel_binding
WHERE platform_account_id = ? AND supplier_goods_no = ? AND is_deleted = 0
`, platformAccountID, supplierGoodsNo); err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, fmt.Errorf("binding not found")
	}
	if len(rows) > 1 {
		return 0, 0, fmt.Errorf("binding duplicated")
	}
	return rows[0].ID, rows[0].GoodsID, nil
}

func (l *TradeOrderLogic) loadGoodsHasTax(ctx context.Context, goodsID int64) (int, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT has_tax FROM product_goods WHERE id = ? AND is_deleted = 0`, goodsID)
	if err != nil {
		return 0, apiErr(consts.CodeInternalError, "读取商品失败")
	}
	if value.IsNil() {
		return 0, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	return value.Int(), nil
}

type bindingCostAdjustResult struct {
	SourceCostPrice    string
	CostPrice          string
	TaxAdjustDirection string
	TaxAdjustRate      string
	TaxAdjustAmount    string
}

func (l *TradeOrderLogic) calcBindingCostPrice(ctx context.Context, goodsHasTax, accountHasTax int, sourceCost decimal.Decimal) (bindingCostAdjustResult, error) {
	if goodsHasTax == accountHasTax {
		return bindingCostAdjustResult{
			SourceCostPrice:    MoneyString(sourceCost),
			CostPrice:          MoneyString(sourceCost),
			TaxAdjustDirection: "none",
			TaxAdjustRate:      "0.0000",
			TaxAdjustAmount:    "0.0000",
		}, nil
	}

	untaxedToTaxed, err := l.getTradeTaxRate(ctx, "trade.tax.untaxed_to_taxed_rate")
	if err != nil {
		return bindingCostAdjustResult{}, err
	}
	taxedToUntaxed, err := l.getTradeTaxRate(ctx, "trade.tax.taxed_to_untaxed_rate")
	if err != nil {
		return bindingCostAdjustResult{}, err
	}
	adjust, err := TaxAdjust(goodsHasTax, accountHasTax, sourceCost, untaxedToTaxed, taxedToUntaxed)
	if err != nil {
		return bindingCostAdjustResult{}, apiErr(consts.CodeBadRequest, "税态参数错误")
	}
	return bindingCostAdjustResult{
		SourceCostPrice:    MoneyString(sourceCost),
		CostPrice:          MoneyString(adjust.CostPrice),
		TaxAdjustDirection: strings.TrimSpace(adjust.TaxAdjustDirection),
		TaxAdjustRate:      MoneyString(adjust.TaxAdjustRate),
		TaxAdjustAmount:    MoneyString(adjust.TaxAdjustAmount),
	}, nil
}

func (l *TradeOrderLogic) getTradeTaxRate(ctx context.Context, key string) (decimal.Decimal, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT config_value FROM system_config WHERE config_key = ?`, key)
	if err != nil {
		return decimal.Zero, apiErr(consts.CodeInternalError, "读取税点失败")
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return decimal.Zero, apiErr(consts.CodeBadRequest, "税点未配置")
	}
	rate, err := decimal.NewFromString(raw)
	if err != nil || rate.IsNegative() {
		return decimal.Zero, apiErr(consts.CodeBadRequest, "税点格式错误")
	}
	return rate, nil
}
