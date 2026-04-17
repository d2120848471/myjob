package adminlogic

import (
	"context"
	"fmt"
	"sort"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

// Create 新增绑定，并计算比较成本价与税态换算字段。
func (l *ProductGoodsChannelBindingLogic) Create(ctx context.Context, req *adminapi.ProductGoodsChannelBindingCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingCreateRes, error) {
	goods, err := l.getActiveGoodsForBinding(ctx, req.GoodsID)
	if err != nil {
		return nil, err
	}
	if goods.SupplyType != "channel" {
		return nil, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}

	account, err := l.getActiveAccountForBinding(ctx, req.PlatformAccountID)
	if err != nil {
		return nil, err
	}
	if account.SubjectID != goods.SubjectID {
		return nil, apiErr(consts.CodeBadRequest, "渠道账号主体必须与商品主体一致")
	}

	supplierGoodsNo, err := normalizeBindingSupplierGoodsNo(req.SupplierGoodsNo)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	supplierGoodsName := normalizeBindingSupplierGoodsName(req.SupplierGoodsName)

	dockStatus, err := normalizeBindingDockStatus(req.DockStatus)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if req.Weight < 0 {
		return nil, apiErr(consts.CodeBadRequest, "weight不能小于0")
	}
	startTime, endTime, err := normalizeBindingTimePeriod(req.StartTime, req.EndTime)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	sourceCostAmount, sourceCostPrice, err := normalizeNonNegativeMoney(req.SourceCostPrice, "source_cost_price")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	if err := l.ensureBindingUnique(ctx, req.GoodsID, 0, req.PlatformAccountID, supplierGoodsNo); err != nil {
		return nil, err
	}

	validateTemplateAny, err := l.normalizeValidateTemplateID(ctx, req.ValidateTemplateID)
	if err != nil {
		return nil, err
	}

	cost, err := l.calcBindingCostPrice(ctx, goods.HasTax, account.HasTax, sourceCostAmount)
	if err != nil {
		return nil, err
	}

	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (
    goods_id,
    platform_account_id,
    supplier_goods_no,
    supplier_goods_name,
    source_cost_price,
    cost_price,
    tax_adjust_direction,
    tax_adjust_rate,
    tax_adjust_amount,
    dock_status,
    sort,
    weight,
    start_time,
    end_time,
    validate_template_id,
    created_at,
    updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, req.GoodsID,
		req.PlatformAccountID,
		supplierGoodsNo,
		supplierGoodsName,
		sourceCostPrice,
		cost.CostPrice,
		cost.TaxAdjustDirection,
		cost.TaxAdjustRate,
		cost.TaxAdjustAmount,
		dockStatus,
		req.Sort,
		req.Weight,
		startTime,
		endTime,
		validateTemplateAny,
		l.core.Now(),
		l.core.Now(),
	)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "新增绑定失败")
	}
	bindingID, _ := result.LastInsertId()

	if err := l.afterBindingChanged(ctx, req.GoodsID, bindingID, supplierGoodsName); err != nil {
		return nil, apiErr(consts.CodeInternalError, "新增绑定失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品渠道绑定：goods_id=%d binding_id=%d platform_account_id=%d supplier_goods_no=%s", req.GoodsID, bindingID, req.PlatformAccountID, supplierGoodsNo), ip)
	return &adminapi.ProductGoodsChannelBindingCreateRes{ID: bindingID}, nil
}

// Update 更新绑定基础字段，并重新计算比较成本价。
func (l *ProductGoodsChannelBindingLogic) Update(ctx context.Context, req *adminapi.ProductGoodsChannelBindingUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingUpdateRes, error) {
	goods, err := l.getActiveGoodsForBinding(ctx, req.GoodsID)
	if err != nil {
		return nil, err
	}
	if goods.SupplyType != "channel" {
		return nil, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}
	if _, err := l.getActiveBindingByID(ctx, req.GoodsID, req.BindingID); err != nil {
		return nil, err
	}

	account, err := l.getActiveAccountForBinding(ctx, req.PlatformAccountID)
	if err != nil {
		return nil, err
	}
	if account.SubjectID != goods.SubjectID {
		return nil, apiErr(consts.CodeBadRequest, "渠道账号主体必须与商品主体一致")
	}

	supplierGoodsNo, err := normalizeBindingSupplierGoodsNo(req.SupplierGoodsNo)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	supplierGoodsName := normalizeBindingSupplierGoodsName(req.SupplierGoodsName)

	dockStatus, err := normalizeBindingDockStatus(req.DockStatus)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if req.Weight < 0 {
		return nil, apiErr(consts.CodeBadRequest, "weight不能小于0")
	}
	startTime, endTime, err := normalizeBindingTimePeriod(req.StartTime, req.EndTime)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	sourceCostAmount, sourceCostPrice, err := normalizeNonNegativeMoney(req.SourceCostPrice, "source_cost_price")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	if err := l.ensureBindingUnique(ctx, req.GoodsID, req.BindingID, req.PlatformAccountID, supplierGoodsNo); err != nil {
		return nil, err
	}

	validateTemplateAny, err := l.normalizeValidateTemplateID(ctx, req.ValidateTemplateID)
	if err != nil {
		return nil, err
	}

	cost, err := l.calcBindingCostPrice(ctx, goods.HasTax, account.HasTax, sourceCostAmount)
	if err != nil {
		return nil, err
	}

	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET platform_account_id = ?,
    supplier_goods_no = ?,
    supplier_goods_name = ?,
    source_cost_price = ?,
    cost_price = ?,
    tax_adjust_direction = ?,
    tax_adjust_rate = ?,
    tax_adjust_amount = ?,
    dock_status = ?,
    sort = ?,
    weight = ?,
    start_time = ?,
    end_time = ?,
    validate_template_id = ?,
    updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, req.PlatformAccountID,
		supplierGoodsNo,
		supplierGoodsName,
		sourceCostPrice,
		cost.CostPrice,
		cost.TaxAdjustDirection,
		cost.TaxAdjustRate,
		cost.TaxAdjustAmount,
		dockStatus,
		req.Sort,
		req.Weight,
		startTime,
		endTime,
		validateTemplateAny,
		l.core.Now(),
		req.BindingID,
		req.GoodsID,
	); err != nil {
		return nil, apiErr(consts.CodeInternalError, "更新绑定失败")
	}

	if err := l.afterBindingChanged(ctx, req.GoodsID, req.BindingID, supplierGoodsName); err != nil {
		return nil, apiErr(consts.CodeInternalError, "更新绑定失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("更新商品渠道绑定：goods_id=%d binding_id=%d platform_account_id=%d supplier_goods_no=%s", req.GoodsID, req.BindingID, req.PlatformAccountID, supplierGoodsNo), ip)
	return &adminapi.ProductGoodsChannelBindingUpdateRes{}, nil
}

// Delete 删除绑定（软删）。
func (l *ProductGoodsChannelBindingLogic) Delete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingDeleteRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	if _, err := l.getActiveBindingByID(ctx, req.GoodsID, req.BindingID); err != nil {
		return nil, err
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_deleted = 1, deleted_at = ?, updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, l.core.Now(), l.core.Now(), req.BindingID, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "删除绑定失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品渠道绑定：goods_id=%d binding_id=%d", req.GoodsID, req.BindingID), ip)
	return &adminapi.ProductGoodsChannelBindingDeleteRes{}, nil
}

// BatchStatus 批量启停绑定。
func (l *ProductGoodsChannelBindingLogic) BatchStatus(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingBatchStatusRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	if len(req.BindingIDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "binding_ids不能为空")
	}
	dockStatus, err := normalizeBindingDockStatus(req.DockStatus)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	args := make([]any, 0, len(req.BindingIDs)+3)
	args = append(args, dockStatus, l.core.Now(), req.GoodsID)
	for _, id := range req.BindingIDs {
		args = append(args, id)
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET dock_status = ?, updated_at = ?
WHERE goods_id = ? AND id IN (`+sqlPlaceholders(len(req.BindingIDs))+`) AND is_deleted = 0
`, args...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "批量启停失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量启停商品渠道绑定：goods_id=%d count=%d dock_status=%s", req.GoodsID, len(req.BindingIDs), dockStatus), ip)
	return &adminapi.ProductGoodsChannelBindingBatchStatusRes{}, nil
}

// BatchDelete 批量删除绑定（软删）。
func (l *ProductGoodsChannelBindingLogic) BatchDelete(ctx context.Context, req *adminapi.ProductGoodsChannelBindingBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingBatchDeleteRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	if len(req.BindingIDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "binding_ids不能为空")
	}
	now := l.core.Now()
	args := make([]any, 0, len(req.BindingIDs)+4)
	args = append(args, now, now, req.GoodsID)
	for _, id := range req.BindingIDs {
		args = append(args, id)
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_deleted = 1, deleted_at = ?, updated_at = ?
WHERE goods_id = ? AND id IN (`+sqlPlaceholders(len(req.BindingIDs))+`) AND is_deleted = 0
`, args...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "批量删除失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量删除商品渠道绑定：goods_id=%d count=%d", req.GoodsID, len(req.BindingIDs)), ip)
	return &adminapi.ProductGoodsChannelBindingBatchDeleteRes{}, nil
}

// Reorder 一键排序绑定。
func (l *ProductGoodsChannelBindingLogic) Reorder(ctx context.Context, req *adminapi.ProductGoodsChannelBindingReorderReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingReorderRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	type row struct {
		ID        int64  `db:"id"`
		CostPrice string `db:"cost_price"`
		Sort      int    `db:"sort"`
	}
	rows := make([]row, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, cost_price, sort
FROM product_goods_channel_binding
WHERE goods_id = ? AND is_deleted = 0
`, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "一键排序失败")
	}
	sort.Slice(rows, func(i, j int) bool {
		left, leftOK := parseMoneyDecimal(rows[i].CostPrice)
		right, rightOK := parseMoneyDecimal(rows[j].CostPrice)
		if leftOK != rightOK {
			return leftOK
		}
		if leftOK && rightOK && !left.Equal(right) {
			return left.LessThan(right)
		}
		if rows[i].Sort != rows[j].Sort {
			return rows[i].Sort < rows[j].Sort
		}
		return rows[i].ID < rows[j].ID
	})

	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for i, item := range rows {
			newSort := (i + 1) * 10
			if _, err := tx.Exec(`UPDATE product_goods_channel_binding SET sort = ?, updated_at = ? WHERE id = ? AND goods_id = ? AND is_deleted = 0`, newSort, l.core.Now(), item.ID, req.GoodsID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "一键排序失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("一键排序商品渠道绑定：goods_id=%d", req.GoodsID), ip)
	return &adminapi.ProductGoodsChannelBindingReorderRes{}, nil
}

// AutoPriceUpdate 更新单条绑定自动改价字段。
func (l *ProductGoodsChannelBindingLogic) AutoPriceUpdate(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	if _, err := l.getActiveBindingByID(ctx, req.GoodsID, req.BindingID); err != nil {
		return nil, err
	}
	normalized, err := normalizeAutoPriceInput(req.AddType, req.DefaultPrice, req.LockPrice, req.SymbolPrice, req.MaxPrice, req.MinPrice)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_auto_change = ?,
    add_type = ?,
    default_price = ?,
    lock_price = ?,
    symbol_price = ?,
    max_price = ?,
    min_price = ?,
    updated_at = ?
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, req.IsAutoChange,
		normalized.AddType,
		normalized.DefaultPrice,
		normalized.LockPrice,
		normalized.SymbolPrice,
		normalized.MaxPrice,
		normalized.MinPrice,
		l.core.Now(),
		req.BindingID,
		req.GoodsID,
	); err != nil {
		return nil, apiErr(consts.CodeInternalError, "更新自动改价失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("更新商品渠道绑定自动改价：goods_id=%d binding_id=%d", req.GoodsID, req.BindingID), ip)
	return &adminapi.ProductGoodsChannelBindingAutoPriceUpdateRes{}, nil
}

// AutoPriceBatch 批量更新自动改价字段。
func (l *ProductGoodsChannelBindingLogic) AutoPriceBatch(ctx context.Context, req *adminapi.ProductGoodsChannelBindingAutoPriceBatchReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsChannelBindingAutoPriceBatchRes, error) {
	if _, err := l.getActiveGoodsForBinding(ctx, req.GoodsID); err != nil {
		return nil, err
	}
	if len(req.BindingIDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "binding_ids不能为空")
	}
	normalized, err := normalizeAutoPriceInput(req.AddType, req.DefaultPrice, req.LockPrice, req.SymbolPrice, req.MaxPrice, req.MinPrice)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	args := make([]any, 0, len(req.BindingIDs)+9)
	args = append(args,
		req.IsAutoChange,
		normalized.AddType,
		normalized.DefaultPrice,
		normalized.LockPrice,
		normalized.SymbolPrice,
		normalized.MaxPrice,
		normalized.MinPrice,
		l.core.Now(),
		req.GoodsID,
	)
	for _, id := range req.BindingIDs {
		args = append(args, id)
	}

	if _, err := l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_binding
SET is_auto_change = ?,
    add_type = ?,
    default_price = ?,
    lock_price = ?,
    symbol_price = ?,
    max_price = ?,
    min_price = ?,
    updated_at = ?
WHERE goods_id = ? AND id IN (`+sqlPlaceholders(len(req.BindingIDs))+`) AND is_deleted = 0
`, args...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "批量更新自动改价失败")
	}
	_ = (&ProductGoodsChannelConfigLogic{core: l.core}).refreshGoodsChannelSummary(ctx, req.GoodsID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量更新商品渠道绑定自动改价：goods_id=%d count=%d", req.GoodsID, len(req.BindingIDs)), ip)
	return &adminapi.ProductGoodsChannelBindingAutoPriceBatchRes{}, nil
}

func (l *ProductGoodsChannelBindingLogic) afterBindingChanged(ctx context.Context, goodsID int64, bindingID int64, supplierGoodsName string) error {
	cfgLogic := &ProductGoodsChannelConfigLogic{core: l.core}
	if err := cfgLogic.refreshGoodsChannelSummary(ctx, goodsID); err != nil {
		return err
	}
	if err := l.updateGoodsNameSnapshotIfNeeded(ctx, goodsID, bindingID, supplierGoodsName); err != nil {
		return err
	}
	return nil
}

func (l *ProductGoodsChannelBindingLogic) updateGoodsNameSnapshotIfNeeded(ctx context.Context, goodsID int64, bindingID int64, supplierGoodsName string) error {
	supplierGoodsName = strings.TrimSpace(supplierGoodsName)
	if supplierGoodsName == "" {
		return nil
	}
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT sync_goods_name_enabled, primary_binding_id
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return err
	}
	if record == nil || len(record) == 0 {
		return nil
	}
	if record["sync_goods_name_enabled"].Int() == 0 {
		return nil
	}
	primary := productGoodsRecordNullInt64(record, "primary_binding_id")
	if !primary.Valid || primary.Int64 != bindingID {
		return nil
	}
	_, err = l.core.DB().Exec(ctx, `UPDATE product_goods SET name = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`, supplierGoodsName, l.core.Now(), goodsID)
	return err
}

func (l *ProductGoodsChannelBindingLogic) getActiveBindingByID(ctx context.Context, goodsID int64, bindingID int64) (gdb.Record, error) {
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT id, platform_account_id, supplier_goods_no
FROM product_goods_channel_binding
WHERE id = ? AND goods_id = ? AND is_deleted = 0
`, bindingID, goodsID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取绑定失败")
	}
	if record == nil || len(record) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "绑定不存在")
	}
	return record, nil
}

func (l *ProductGoodsChannelBindingLogic) normalizeValidateTemplateID(ctx context.Context, id *int64) (any, error) {
	if id == nil || *id <= 0 {
		return nil, nil
	}
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_template WHERE id = ?`, *id)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "校验模板失败")
	}
	if exists.Int() == 0 {
		return nil, apiErr(consts.CodeBadRequest, "模板不存在")
	}
	return *id, nil
}

type autoPriceInput struct {
	AddType      string
	DefaultPrice string
	LockPrice    string
	SymbolPrice  string
	MaxPrice     string
	MinPrice     string
}

func normalizeAutoPriceInput(addType string, defaultPrice string, lockPrice string, symbolPrice string, maxPrice string, minPrice string) (autoPriceInput, error) {
	addType = strings.TrimSpace(addType)
	if addType == "" {
		addType = "fixed"
	}
	if addType != "fixed" && addType != "percent" {
		return autoPriceInput{}, fmt.Errorf("add_type错误")
	}
	defaultMoney, err := normalizeMoneyWithMin(defaultPrice, decimal.NewFromInt(-1))
	if err != nil {
		return autoPriceInput{}, fmt.Errorf("default_price格式错误")
	}
	lockMoney, err := normalizeMoneyWithMin(lockPrice, decimal.Zero)
	if err != nil {
		return autoPriceInput{}, fmt.Errorf("lock_price格式错误")
	}
	symbolMoney, err := normalizeMoneyWithMin(symbolPrice, decimal.Zero)
	if err != nil {
		return autoPriceInput{}, fmt.Errorf("symbol_price格式错误")
	}
	maxMoney, err := normalizeMoneyWithMin(maxPrice, decimal.Zero)
	if err != nil {
		return autoPriceInput{}, fmt.Errorf("max_price格式错误")
	}
	minMoney, err := normalizeMoneyWithMin(minPrice, decimal.Zero)
	if err != nil {
		return autoPriceInput{}, fmt.Errorf("min_price格式错误")
	}
	return autoPriceInput{
		AddType:      addType,
		DefaultPrice: defaultMoney,
		LockPrice:    lockMoney,
		SymbolPrice:  symbolMoney,
		MaxPrice:     maxMoney,
		MinPrice:     minMoney,
	}, nil
}

func normalizeMoneyWithMin(value string, minValue decimal.Decimal) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0.0000", nil
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return "", err
	}
	if amount.LessThan(minValue) {
		return "", fmt.Errorf("money too small")
	}
	return amount.StringFixed(4), nil
}
