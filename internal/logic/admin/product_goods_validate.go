package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/shopspring/decimal"
)

type normalizedProductGoodsInput struct {
	BrandID                 int64
	Name                    string
	GoodsType               string
	SupplyType              string
	IsExport                int
	IsDouyin                int
	HasTax                  int
	SubjectID               *int64
	ExceptionNotify         int
	ProductTemplateID       *int64
	PurchaseLimitStrategyID *int64
	PurchaseNotice          string
	TerminalPriceLimit      string
	BalanceLimit            string
	DefaultSellPrice        string
	MinPurchaseQty          int
	MaxPurchaseQty          int
	Status                  int
}

func (l *ProductGoodsLogic) normalizeProductGoodsInput(ctx context.Context, brandID int64, name, goodsType, supplyType string, isExport, isDouyin, hasTax int, subjectID *int64, exceptionNotify int, productTemplateID, purchaseLimitStrategyID *int64, purchaseNotice, terminalPriceLimit, balanceLimit, defaultSellPrice string, minPurchaseQty, maxPurchaseQty, status int, allowDisabledCurrentStrategyID *int64) (normalizedProductGoodsInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "商品名称不能为空")
	}
	if _, ok := productGoodsTypeLabels[strings.TrimSpace(goodsType)]; !ok {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "商品类型错误")
	}
	supplyType = strings.TrimSpace(strings.ToLower(supplyType))
	if supplyType != productGoodsSupplyTypeChannel {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "供货方式错误")
	}
	if err := validateBooleanFlag(isExport, "可导出"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(isDouyin, "可抖音"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(hasTax, "含税标识"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if err := validateBooleanFlag(exceptionNotify, "异常提醒"); err != nil {
		return normalizedProductGoodsInput{}, err
	}
	if status != consts.StatusEnabled && status != consts.StatusDisabled {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if _, err := l.validateLeafBrand(ctx, brandID); err != nil {
		return normalizedProductGoodsInput{}, err
	}

	normalizedSubjectID := normalizeOptionalID(subjectID)
	if hasTax == 1 {
		if normalizedSubjectID == nil {
			return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "含税商品必须选择主体")
		}
		// 含税商品后续会走开票链路，这里提前锁死主体存在且可开票。
		if err := l.ensureTaxSubjectUsable(ctx, *normalizedSubjectID); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	} else {
		// 不含税商品不保留主体，避免后续出现脏数据。
		normalizedSubjectID = nil
	}

	normalizedTemplateID := normalizeOptionalID(productTemplateID)
	if normalizedTemplateID != nil {
		if err := l.ensureTemplateExists(ctx, *normalizedTemplateID); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	}

	normalizedStrategyID := normalizeOptionalID(purchaseLimitStrategyID)
	if normalizedStrategyID != nil {
		allowDisabled := allowDisabledCurrentStrategyID != nil && *allowDisabledCurrentStrategyID == *normalizedStrategyID
		if err := l.ensureStrategyUsable(ctx, *normalizedStrategyID, allowDisabled); err != nil {
			return normalizedProductGoodsInput{}, err
		}
	}

	normalizedTerminalPriceLimit, err := normalizeOptionalMoney(terminalPriceLimit)
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "终端限价格式错误")
	}
	normalizedBalanceLimit, err := normalizeDefaultMoney(balanceLimit, "0")
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "余额限制格式错误")
	}
	normalizedDefaultSellPrice, err := normalizeOptionalMoney(defaultSellPrice)
	if err != nil {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "默认售价格式错误")
	}
	if minPurchaseQty < 1 {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "最小购买数量必须大于等于1")
	}
	if maxPurchaseQty < minPurchaseQty {
		return normalizedProductGoodsInput{}, apiErr(consts.CodeBadRequest, "最大购买数量不能小于最小购买数量")
	}

	return normalizedProductGoodsInput{
		BrandID:                 brandID,
		Name:                    name,
		GoodsType:               strings.TrimSpace(goodsType),
		SupplyType:              supplyType,
		IsExport:                isExport,
		IsDouyin:                isDouyin,
		HasTax:                  hasTax,
		SubjectID:               normalizedSubjectID,
		ExceptionNotify:         exceptionNotify,
		ProductTemplateID:       normalizedTemplateID,
		PurchaseLimitStrategyID: normalizedStrategyID,
		PurchaseNotice:          strings.TrimSpace(purchaseNotice),
		TerminalPriceLimit:      normalizedTerminalPriceLimit,
		BalanceLimit:            normalizedBalanceLimit,
		DefaultSellPrice:        normalizedDefaultSellPrice,
		MinPurchaseQty:          minPurchaseQty,
		MaxPurchaseQty:          maxPurchaseQty,
		Status:                  status,
	}, nil
}

func (l *ProductGoodsLogic) validateLeafBrand(ctx context.Context, brandID int64) (entity.ProductBrand, error) {
	if brandID <= 0 {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌不能为空")
	}
	brand := entity.ProductBrand{}
	if err := l.core.DB().GetCore().GetScan(ctx, &brand, `SELECT id, parent_id, name, icon, credential_image, COALESCE(description, '') AS description, is_visible, sort, goods_count, created_at, updated_at FROM product_brand WHERE id = ?`, brandID); err != nil {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌不存在")
	}
	childCount, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_brand WHERE parent_id = ?`, brandID)
	if err != nil {
		return entity.ProductBrand{}, apiErr(consts.CodeInternalError, "品牌校验失败")
	}
	if childCount.Int() > 0 {
		return entity.ProductBrand{}, apiErr(consts.CodeBadRequest, "品牌必须选择末级品牌")
	}
	return brand, nil
}

func (l *ProductGoodsLogic) ensureTemplateExists(ctx context.Context, id int64) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_template WHERE id = ?`, id)
	if err != nil {
		return apiErr(consts.CodeInternalError, "商品模板校验失败")
	}
	if count.Int() == 0 {
		return apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	return nil
}

func (l *ProductGoodsLogic) ensureTaxSubjectUsable(ctx context.Context, id int64) error {
	row := struct {
		ID     int64 `db:"id"`
		HasTax int   `db:"has_tax"`
	}{}
	if err := l.core.DB().GetCore().GetScan(ctx, &row, `SELECT id, has_tax FROM admin_subject WHERE id = ?`, id); err != nil {
		return apiErr(consts.CodeBadRequest, "主体不存在")
	}
	if row.HasTax != 1 {
		return apiErr(consts.CodeBadRequest, "含税商品必须选择含税主体")
	}
	return nil
}

func (l *ProductGoodsLogic) ensureStrategyUsable(ctx context.Context, id int64, allowDisabled bool) error {
	row := struct {
		ID     int64 `db:"id"`
		Status int   `db:"status"`
	}{}
	if err := l.core.DB().GetCore().GetScan(ctx, &row, `SELECT id, status FROM product_purchase_limit_strategy WHERE id = ?`, id); err != nil {
		return apiErr(consts.CodeBadRequest, "购买数量限制策略不存在")
	}
	if row.Status != consts.StatusEnabled && !allowDisabled {
		return apiErr(consts.CodeBadRequest, "购买数量限制策略必须为启用状态")
	}
	return nil
}

func normalizeProductGoodsTriState(value string, fieldName string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("%s筛选错误", fieldName)
	}
}

func validateBooleanFlag(value int, fieldName string) error {
	if value != 0 && value != 1 {
		return apiErr(consts.CodeBadRequest, fieldName+"错误")
	}
	return nil
}

func normalizeOptionalID(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	id := *value
	return &id
}

func normalizeOptionalMoney(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.IsNegative() {
		return "", fmt.Errorf("money format invalid")
	}
	return amount.StringFixed(4), nil
}

func normalizeDefaultMoney(value string, defaultValue string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = defaultValue
	}
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.IsNegative() {
		return "", fmt.Errorf("money format invalid")
	}
	return amount.StringFixed(4), nil
}
