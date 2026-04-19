package adminlogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"

	"github.com/shopspring/decimal"
)

type normalizedProductGoodsChannelBindingInput struct {
	PlatformAccountID  int64
	SupplierGoodsNo    string
	SupplierGoodsName  string
	SourceCostPrice    string
	CostSnapshot       productGoodsChannelCostSnapshot
	ValidateTemplateID *int64
	DockStatus         int
	Sort               int
	OrderWeight        string
	OrderTimeStart     string
	OrderTimeEnd       string
}

type normalizedProductGoodsChannelAutoPriceInput struct {
	IsAutoChange int
	AddType      string
	DefaultPrice string
}

func (l *ProductGoodsLogic) normalizeProductGoodsChannelBindingInput(ctx context.Context, goods entity.ProductGoods, platformAccountID int64, supplierGoodsNo, supplierGoodsName, sourceCostPrice string, validateTemplateID *int64, dockStatus, sort int, orderWeight, orderTimeStart, orderTimeEnd string, currentBindingID *int64) (normalizedProductGoodsChannelBindingInput, error) {
	if goods.ID <= 0 || goods.IsDeleted != 0 {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	if goods.SupplyType != productGoodsSupplyTypeChannel {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "仅渠道供货商品允许维护渠道绑定")
	}
	if platformAccountID <= 0 {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "渠道账号不能为空")
	}

	account, err := l.getActiveSupplierPlatformAccount(ctx, platformAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "渠道账号不存在")
		}
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeInternalError, "渠道账号校验失败")
	}
	if account.Status != consts.StatusEnabled {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "渠道账号已关闭")
	}

	if goods.SubjectID.Valid && goods.SubjectID.Int64 != account.SubjectID {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "渠道账号主体必须与商品主体一致")
	}

	supplierGoodsNo = strings.TrimSpace(supplierGoodsNo)
	if supplierGoodsNo == "" {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "对接商品编号不能为空")
	}
	supplierGoodsName = strings.TrimSpace(supplierGoodsName)
	if supplierGoodsName == "" {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "对接商品名称不能为空")
	}

	normalizedSourceCostPrice, err := normalizeOptionalMoney(sourceCostPrice)
	if err != nil || normalizedSourceCostPrice == "" {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "原始进货价格式错误")
	}

	normalizedTemplateID := normalizeOptionalID(validateTemplateID)
	if normalizedTemplateID != nil {
		if err := l.ensureTemplateExists(ctx, *normalizedTemplateID); err != nil {
			return normalizedProductGoodsChannelBindingInput{}, err
		}
	}

	if err := validateBooleanFlag(dockStatus, "对接状态"); err != nil {
		return normalizedProductGoodsChannelBindingInput{}, err
	}
	if sort < 0 {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "排序值不能小于0")
	}
	normalizedOrderWeight, err := normalizeDefaultMoney(orderWeight, "0")
	if err != nil {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "下单权重格式错误")
	}
	normalizedOrderTimeStart, normalizedOrderTimeEnd, err := normalizeOrderTimeWindow(orderTimeStart, orderTimeEnd)
	if err != nil {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, err.Error())
	}

	if err := l.ensureProductGoodsChannelBindingUnique(ctx, goods.ID, platformAccountID, supplierGoodsNo, currentBindingID); err != nil {
		return normalizedProductGoodsChannelBindingInput{}, err
	}

	financeTaxConfig, err := l.loadProductGoodsChannelFinanceTaxConfig(ctx, goods.HasTax, account.HasTax)
	if err != nil {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, "税率未配置")
	}
	costSnapshot, err := computeChannelCostSnapshot(normalizedSourceCostPrice, goods.HasTax, account.HasTax, financeTaxConfig)
	if err != nil {
		return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeBadRequest, err.Error())
	}

	normalizedSort := sort
	if currentBindingID == nil && normalizedSort == 0 {
		normalizedSort, err = l.nextProductGoodsChannelBindingSort(ctx, goods.ID)
		if err != nil {
			return normalizedProductGoodsChannelBindingInput{}, apiErr(consts.CodeInternalError, "绑定排序计算失败")
		}
	}

	return normalizedProductGoodsChannelBindingInput{
		PlatformAccountID:  platformAccountID,
		SupplierGoodsNo:    supplierGoodsNo,
		SupplierGoodsName:  supplierGoodsName,
		SourceCostPrice:    normalizedSourceCostPrice,
		CostSnapshot:       costSnapshot,
		ValidateTemplateID: normalizedTemplateID,
		DockStatus:         dockStatus,
		Sort:               normalizedSort,
		OrderWeight:        normalizedOrderWeight,
		OrderTimeStart:     normalizedOrderTimeStart,
		OrderTimeEnd:       normalizedOrderTimeEnd,
	}, nil
}

func (l *ProductGoodsLogic) normalizeProductGoodsChannelAutoPriceInput(isAutoChange int, addType, defaultPrice string) (normalizedProductGoodsChannelAutoPriceInput, error) {
	if err := validateBooleanFlag(isAutoChange, "自动改价状态"); err != nil {
		return normalizedProductGoodsChannelAutoPriceInput{}, err
	}
	if isAutoChange == 0 {
		return normalizedProductGoodsChannelAutoPriceInput{
			IsAutoChange: 0,
			AddType:      "",
			DefaultPrice: "0.0000",
		}, nil
	}

	addType = strings.TrimSpace(strings.ToLower(addType))
	if addType != autoPriceAddTypeFixed && addType != autoPriceAddTypePercent {
		return normalizedProductGoodsChannelAutoPriceInput{}, apiErr(consts.CodeBadRequest, "自动改价类型错误")
	}
	normalizedDefaultPrice, err := normalizeDefaultMoney(defaultPrice, "0")
	if err != nil {
		return normalizedProductGoodsChannelAutoPriceInput{}, apiErr(consts.CodeBadRequest, "利润值格式错误")
	}
	return normalizedProductGoodsChannelAutoPriceInput{
		IsAutoChange: isAutoChange,
		AddType:      addType,
		DefaultPrice: normalizedDefaultPrice,
	}, nil
}

func (l *ProductGoodsLogic) loadProductGoodsChannelFinanceTaxConfig(ctx context.Context, goodsHasTax, channelHasTax int) (modelruntime.FinanceTaxConfig, error) {
	if goodsHasTax == channelHasTax {
		return modelruntime.FinanceTaxConfig{}, nil
	}

	state, err := l.core.LoadSystemConfigGroup(ctx, "finance")
	if err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}

	items := make(map[string]modelruntime.SystemConfigItem, len(state.Items))
	for _, item := range state.Items {
		items[item.Key] = item
	}

	if goodsHasTax == 1 && channelHasTax == 0 {
		rate, rateErr := normalizeRequiredFinanceRate(items, "tax_exclusive_rate", "未税->含税税率")
		if rateErr != nil {
			return modelruntime.FinanceTaxConfig{}, rateErr
		}
		return modelruntime.FinanceTaxConfig{TaxExclusiveRate: rate}, nil
	}

	rate, rateErr := normalizeRequiredFinanceRate(items, "tax_inclusive_rate", "含税->未税税率")
	if rateErr != nil {
		return modelruntime.FinanceTaxConfig{}, rateErr
	}
	return modelruntime.FinanceTaxConfig{TaxInclusiveRate: rate}, nil
}

func normalizeRequiredFinanceRate(items map[string]modelruntime.SystemConfigItem, key, label string) (string, error) {
	item, ok := items[key]
	if !ok || !item.Configured {
		return "", app.ErrSystemConfigNotConfigured
	}

	value := strings.TrimSpace(item.Value)
	if value == "" {
		return "", app.ErrSystemConfigNotConfigured
	}
	rate, err := decimal.NewFromString(value)
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return "", apiErr(consts.CodeBadRequest, label+"格式错误")
	}
	return rate.StringFixed(4), nil
}

func normalizeOrderTimeWindow(start, end string) (string, string, error) {
	start = strings.TrimSpace(start)
	end = strings.TrimSpace(end)
	if start == "" && end == "" {
		return "", "", nil
	}
	if start == "" || end == "" {
		return "", "", fmt.Errorf("下单时段必须同时填写开始和结束时间")
	}
	if _, err := time.Parse("15:04", start); err != nil {
		return "", "", fmt.Errorf("下单开始时段格式错误")
	}
	if _, err := time.Parse("15:04", end); err != nil {
		return "", "", fmt.Errorf("下单结束时段格式错误")
	}
	return start, end, nil
}

func (l *ProductGoodsLogic) getActiveSupplierPlatformAccount(ctx context.Context, id int64) (entity.SupplierPlatformAccount, error) {
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    id, name, provider_code, provider_name, type_id, subject_id, has_tax, status, domain, backup_domain,
    token_id, secret_key, extra_config, threshold_amount, sort, crowd_name, last_balance,
    last_balance_status, last_balance_message, last_balance_at, last_balance_trace_id, is_deleted,
    deleted_at, created_at, updated_at
FROM supplier_platform_account
WHERE id = ? AND is_deleted = 0
`, id)
	if err != nil {
		return entity.SupplierPlatformAccount{}, err
	}
	if len(rows) == 0 {
		return entity.SupplierPlatformAccount{}, sql.ErrNoRows
	}
	return supplierPlatformAccountFromRecord(rows[0]), nil
}

func (l *ProductGoodsLogic) ensureProductGoodsChannelBindingUnique(ctx context.Context, goodsID, platformAccountID int64, supplierGoodsNo string, currentBindingID *int64) error {
	query := `
SELECT COUNT(*)
FROM product_goods_channel_binding
WHERE goods_id = ? AND platform_account_id = ? AND supplier_goods_no = ? AND is_deleted = 0
`
	args := []any{goodsID, platformAccountID, supplierGoodsNo}
	if currentBindingID != nil && *currentBindingID > 0 {
		query += ` AND id <> ?`
		args = append(args, *currentBindingID)
	}
	count, err := l.core.DB().GetCore().GetValue(ctx, query, args...)
	if err != nil {
		return apiErr(consts.CodeInternalError, "渠道绑定校验失败")
	}
	if count.Int() > 0 {
		return apiErr(consts.CodeBadRequest, "同一商品下渠道账号和对接商品编号不能重复")
	}
	return nil
}

func (l *ProductGoodsLogic) nextProductGoodsChannelBindingSort(ctx context.Context, goodsID int64) (int, error) {
	value, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COALESCE(MAX(sort), 0) FROM product_goods_channel_binding WHERE goods_id = ? AND is_deleted = 0`, goodsID)
	if err != nil {
		return 0, err
	}
	return value.Int() + 1, nil
}
