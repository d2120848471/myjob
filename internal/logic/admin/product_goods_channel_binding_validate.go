package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"myjob/internal/consts"

	"github.com/shopspring/decimal"
)

type bindingGoodsRow struct {
	ID          int64
	SupplyType  string
	HasTax      int
	SubjectID   int64
	SubjectName string
}

type bindingAccountRow struct {
	ID           int64
	Name         string
	ProviderCode string
	ProviderName string
	SubjectID    int64
	HasTax       int
}

func normalizeBindingDockStatus(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "enabled", nil
	}
	switch value {
	case "enabled", "disabled":
		return value, nil
	default:
		return "", fmt.Errorf("dock_status错误")
	}
}

func normalizeBindingSupplierGoodsNo(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("supplier_goods_no不能为空")
	}
	return value, nil
}

func normalizeBindingSupplierGoodsName(value string) string {
	return strings.TrimSpace(value)
}

func normalizeBindingTimePeriod(startTime, endTime string) (string, string, error) {
	startTime = strings.TrimSpace(startTime)
	endTime = strings.TrimSpace(endTime)
	if startTime == "" && endTime == "" {
		return "", "", nil
	}
	if startTime == "" || endTime == "" {
		return "", "", fmt.Errorf("配置时段时start_time和end_time必须同时存在")
	}
	if _, ok := parseTimeHM(startTime); !ok {
		return "", "", fmt.Errorf("start_time格式错误")
	}
	if _, ok := parseTimeHM(endTime); !ok {
		return "", "", fmt.Errorf("end_time格式错误")
	}
	return startTime, endTime, nil
}

func normalizeNonNegativeMoney(value string, fieldName string) (decimal.Decimal, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, "", fmt.Errorf("%s不能为空", fieldName)
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, "", fmt.Errorf("%s格式错误", fieldName)
	}
	if amount.IsNegative() {
		return decimal.Zero, "", fmt.Errorf("%s不能小于0", fieldName)
	}
	return amount, amount.StringFixed(4), nil
}

func (l *ProductGoodsChannelBindingLogic) getActiveGoodsForBinding(ctx context.Context, goodsID int64) (bindingGoodsRow, error) {
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    p.id,
    p.supply_type,
    p.has_tax,
    p.subject_id,
    COALESCE(sub.name, '') AS subject_name
FROM product_goods p
LEFT JOIN admin_subject sub ON sub.id = p.subject_id
WHERE p.id = ? AND p.is_deleted = 0
`, goodsID)
	if err != nil {
		return bindingGoodsRow{}, apiErr(consts.CodeInternalError, "读取商品失败")
	}
	if record == nil || len(record) == 0 {
		return bindingGoodsRow{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	subject := productGoodsRecordNullInt64(record, "subject_id")
	if !subject.Valid || subject.Int64 <= 0 {
		return bindingGoodsRow{}, apiErr(consts.CodeBadRequest, "商品主体未配置")
	}
	return bindingGoodsRow{
		ID:          record["id"].Int64(),
		SupplyType:  record["supply_type"].String(),
		HasTax:      record["has_tax"].Int(),
		SubjectID:   subject.Int64,
		SubjectName: productGoodsRecordString(record, "subject_name"),
	}, nil
}

func (l *ProductGoodsChannelBindingLogic) getActiveAccountForBinding(ctx context.Context, accountID int64) (bindingAccountRow, error) {
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT id, name, provider_code, provider_name, subject_id, has_tax
FROM supplier_platform_account
WHERE id = ? AND is_deleted = 0
`, accountID)
	if err != nil {
		return bindingAccountRow{}, apiErr(consts.CodeInternalError, "读取渠道账号失败")
	}
	if record == nil || len(record) == 0 {
		return bindingAccountRow{}, apiErr(consts.CodeBadRequest, "渠道账号不存在")
	}
	return bindingAccountRow{
		ID:           record["id"].Int64(),
		Name:         record["name"].String(),
		ProviderCode: record["provider_code"].String(),
		ProviderName: record["provider_name"].String(),
		SubjectID:    record["subject_id"].Int64(),
		HasTax:       record["has_tax"].Int(),
	}, nil
}

func (l *ProductGoodsChannelBindingLogic) ensureBindingUnique(ctx context.Context, goodsID, bindingID, platformAccountID int64, supplierGoodsNo string) error {
	args := []any{goodsID, platformAccountID, supplierGoodsNo}
	where := "goods_id = ? AND platform_account_id = ? AND supplier_goods_no = ? AND is_deleted = 0"
	if bindingID > 0 {
		where += " AND id <> ?"
		args = append(args, bindingID)
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_binding WHERE `+where, args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "绑定唯一性校验失败")
	}
	if exists.Int() > 0 {
		return apiErr(consts.CodeConflict, "同商品下渠道账号与上游商品编号不允许重复")
	}
	return nil
}

type bindingCostResult struct {
	CostPrice          string
	TaxAdjustDirection string
	TaxAdjustRate      string
	TaxAdjustAmount    string
}

func (l *ProductGoodsChannelBindingLogic) calcBindingCostPrice(ctx context.Context, goodsHasTax, accountHasTax int, sourceCost decimal.Decimal) (bindingCostResult, error) {
	if goodsHasTax == accountHasTax {
		return bindingCostResult{
			CostPrice:          sourceCost.StringFixed(4),
			TaxAdjustDirection: "none",
			TaxAdjustRate:      "0.0000",
			TaxAdjustAmount:    "0.0000",
		}, nil
	}

	if goodsHasTax == 1 && accountHasTax == 0 {
		rate, err := l.getTradeTaxRate(ctx, "trade.tax.untaxed_to_taxed_rate")
		if err != nil {
			return bindingCostResult{}, err
		}
		adjust := sourceCost.Mul(rate).Div(decimal.NewFromInt(100))
		cost := sourceCost.Add(adjust)
		return bindingCostResult{
			CostPrice:          cost.StringFixed(4),
			TaxAdjustDirection: "untaxed_to_taxed",
			TaxAdjustRate:      rate.StringFixed(4),
			TaxAdjustAmount:    adjust.StringFixed(4),
		}, nil
	}

	if goodsHasTax == 0 && accountHasTax == 1 {
		rate, err := l.getTradeTaxRate(ctx, "trade.tax.taxed_to_untaxed_rate")
		if err != nil {
			return bindingCostResult{}, err
		}
		adjust := sourceCost.Mul(rate).Div(decimal.NewFromInt(100))
		cost := sourceCost.Sub(adjust)
		return bindingCostResult{
			CostPrice:          cost.StringFixed(4),
			TaxAdjustDirection: "taxed_to_untaxed",
			TaxAdjustRate:      rate.StringFixed(4),
			TaxAdjustAmount:    adjust.StringFixed(4),
		}, nil
	}

	// 兜底：遇到未知税态值时直接拒绝保存，避免写入错误成本价。
	return bindingCostResult{}, apiErr(consts.CodeBadRequest, "税态参数错误")
}

func (l *ProductGoodsChannelBindingLogic) getTradeTaxRate(ctx context.Context, key string) (decimal.Decimal, error) {
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
