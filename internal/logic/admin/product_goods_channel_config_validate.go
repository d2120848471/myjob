package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/shopspring/decimal"
)

const (
	productGoodsOrderStrategyFixedOrder      = "fixed_order"
	productGoodsOrderStrategyLowestCost      = "lowest_cost"
	productGoodsOrderStrategyWeightedPercent = "weighted_percent"
	productGoodsOrderStrategyTimeWindow      = "time_window"
	productGoodsOrderStrategyRandom          = "random"
)

var productGoodsOrderStrategyLabels = map[string]string{
	productGoodsOrderStrategyFixedOrder:      "固定顺序",
	productGoodsOrderStrategyLowestCost:      "进价从低到高",
	productGoodsOrderStrategyWeightedPercent: "百分比分配",
	productGoodsOrderStrategyTimeWindow:      "按时段提交",
	productGoodsOrderStrategyRandom:          "随机选择",
}

var productGoodsOrderStrategyOptionOrder = []string{
	productGoodsOrderStrategyFixedOrder,
	productGoodsOrderStrategyLowestCost,
	productGoodsOrderStrategyWeightedPercent,
	productGoodsOrderStrategyTimeWindow,
	productGoodsOrderStrategyRandom,
}

var productGoodsOrderWeightTotal = decimal.RequireFromString("100.0000")

type normalizedProductGoodsInventoryConfigInput struct {
	SmartReorderEnabled   int
	ReorderTimeoutEnabled int
	ReorderTimeoutMinutes int
	OrderStrategy         string
	SyncCostPriceEnabled  int
	SyncGoodsNameEnabled  int
	AllowLossSaleEnabled  int
	MaxLossAmount         string
	ComboGoodsEnabled     int
}

type productGoodsInventoryConfigState struct {
	SmartReorderEnabled   int
	ReorderTimeoutEnabled int
	ReorderTimeoutMinutes int
	OrderStrategy         string
	SyncCostPriceEnabled  int
	SyncGoodsNameEnabled  int
	AllowLossSaleEnabled  int
	MaxLossAmount         string
	ComboGoodsEnabled     int
}

func defaultProductGoodsInventoryConfigState() productGoodsInventoryConfigState {
	return productGoodsInventoryConfigState{
		SmartReorderEnabled:   0,
		ReorderTimeoutEnabled: 0,
		ReorderTimeoutMinutes: 0,
		OrderStrategy:         productGoodsOrderStrategyFixedOrder,
		SyncCostPriceEnabled:  0,
		SyncGoodsNameEnabled:  0,
		AllowLossSaleEnabled:  0,
		MaxLossAmount:         "0.0000",
		ComboGoodsEnabled:     0,
	}
}

func (s productGoodsInventoryConfigState) toAPIConfig() adminapi.ProductGoodsInventoryConfig {
	return adminapi.ProductGoodsInventoryConfig{
		SmartReorderEnabled:   s.SmartReorderEnabled,
		ReorderTimeoutEnabled: s.ReorderTimeoutEnabled,
		ReorderTimeoutMinutes: s.ReorderTimeoutMinutes,
		OrderStrategy:         s.OrderStrategy,
		SyncCostPriceEnabled:  s.SyncCostPriceEnabled,
		SyncGoodsNameEnabled:  s.SyncGoodsNameEnabled,
		AllowLossSaleEnabled:  s.AllowLossSaleEnabled,
		MaxLossAmount:         s.MaxLossAmount,
		ComboGoodsEnabled:     s.ComboGoodsEnabled,
	}
}

func (s productGoodsInventoryConfigState) toAPISummary() adminapi.ProductGoodsInventoryConfigSummary {
	return adminapi.ProductGoodsInventoryConfigSummary{
		SmartReorderEnabled:   s.SmartReorderEnabled,
		ReorderTimeoutEnabled: s.ReorderTimeoutEnabled,
		ReorderTimeoutMinutes: s.ReorderTimeoutMinutes,
		OrderStrategy:         s.OrderStrategy,
		SyncCostPriceEnabled:  s.SyncCostPriceEnabled,
		SyncGoodsNameEnabled:  s.SyncGoodsNameEnabled,
		AllowLossSaleEnabled:  s.AllowLossSaleEnabled,
		MaxLossAmount:         s.MaxLossAmount,
		ComboGoodsEnabled:     s.ComboGoodsEnabled,
	}
}

func productGoodsOrderStrategyOptions() []adminapi.ProductGoodsStringOption {
	options := make([]adminapi.ProductGoodsStringOption, 0, len(productGoodsOrderStrategyOptionOrder))
	for _, value := range productGoodsOrderStrategyOptionOrder {
		options = append(options, adminapi.ProductGoodsStringOption{
			Value: value,
			Label: productGoodsOrderStrategyLabels[value],
		})
	}
	return options
}

func ensureChannelSupplyProduct(goods entity.ProductGoods) error {
	if goods.ID <= 0 || goods.IsDeleted != 0 {
		return apiErr(consts.CodeBadRequest, "商品不存在")
	}
	if goods.SupplyType != productGoodsSupplyTypeChannel {
		return apiErr(consts.CodeBadRequest, "仅渠道供货商品允许维护库存配置")
	}
	return nil
}

func (l *ProductGoodsLogic) normalizeProductGoodsInventoryConfigInput(ctx context.Context, goods entity.ProductGoods, req *adminapi.ProductGoodsInventoryConfigSaveReq) (normalizedProductGoodsInventoryConfigInput, error) {
	if err := ensureChannelSupplyProduct(goods); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.SmartReorderEnabled, "智能补单状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.ReorderTimeoutEnabled, "补单超时状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.SyncCostPriceEnabled, "同步进价状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.SyncGoodsNameEnabled, "同步商品名称状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.AllowLossSaleEnabled, "亏本销售状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}
	if err := validateBooleanFlag(req.ComboGoodsEnabled, "组合商品状态"); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}

	orderStrategy := strings.TrimSpace(strings.ToLower(req.OrderStrategy))
	if _, ok := productGoodsOrderStrategyLabels[orderStrategy]; !ok {
		return normalizedProductGoodsInventoryConfigInput{}, apiErr(consts.CodeBadRequest, "下单方式错误")
	}

	reorderTimeoutMinutes := req.ReorderTimeoutMinutes
	if req.ReorderTimeoutEnabled == 0 {
		reorderTimeoutMinutes = 0
	} else if reorderTimeoutMinutes < 1 || reorderTimeoutMinutes > 1440 {
		return normalizedProductGoodsInventoryConfigInput{}, apiErr(consts.CodeBadRequest, "补单时间必须在1到1440分钟之间")
	}

	maxLossAmount, err := normalizeDefaultMoney(req.MaxLossAmount, "0")
	if err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, apiErr(consts.CodeBadRequest, "允许亏本金额格式错误")
	}
	if req.AllowLossSaleEnabled == 0 {
		maxLossAmount = "0.0000"
	}

	if err := l.validateProductGoodsInventoryStrategyBindings(ctx, goods.ID, orderStrategy); err != nil {
		return normalizedProductGoodsInventoryConfigInput{}, err
	}

	return normalizedProductGoodsInventoryConfigInput{
		SmartReorderEnabled:   req.SmartReorderEnabled,
		ReorderTimeoutEnabled: req.ReorderTimeoutEnabled,
		ReorderTimeoutMinutes: reorderTimeoutMinutes,
		OrderStrategy:         orderStrategy,
		SyncCostPriceEnabled:  req.SyncCostPriceEnabled,
		SyncGoodsNameEnabled:  req.SyncGoodsNameEnabled,
		AllowLossSaleEnabled:  req.AllowLossSaleEnabled,
		MaxLossAmount:         maxLossAmount,
		ComboGoodsEnabled:     req.ComboGoodsEnabled,
	}, nil
}

func (l *ProductGoodsLogic) validateProductGoodsInventoryStrategyBindings(ctx context.Context, goodsID int64, orderStrategy string) error {
	switch orderStrategy {
	case productGoodsOrderStrategyWeightedPercent:
		return l.validateProductGoodsWeightedBindings(ctx, goodsID)
	case productGoodsOrderStrategyTimeWindow:
		return l.validateProductGoodsTimeWindowBindings(ctx, goodsID)
	default:
		return nil
	}
}

func (l *ProductGoodsLogic) validateProductGoodsWeightedBindings(ctx context.Context, goodsID int64) error {
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT b.order_weight
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.goods_id = ? AND b.is_deleted = 0 AND b.dock_status = 1 AND a.is_deleted = 0 AND a.status = 1
`, goodsID)
	if err != nil {
		return apiErr(consts.CodeInternalError, "库存配置校验失败")
	}

	sum := decimal.Zero
	validCount := 0
	for _, row := range rows {
		weight, parseErr := decimal.NewFromString(productGoodsRecordMoney(row, "order_weight"))
		if parseErr != nil || weight.IsNegative() {
			continue
		}
		if weight.GreaterThan(decimal.Zero) {
			validCount++
			sum = sum.Add(weight)
		}
	}
	if validCount == 0 {
		return apiErr(consts.CodeBadRequest, "百分比分配至少需要一条权重大于0的启用绑定")
	}
	if !sum.Equal(productGoodsOrderWeightTotal) {
		return apiErr(consts.CodeBadRequest, "百分比分配的权重总和必须等于100")
	}
	return nil
}

func (l *ProductGoodsLogic) validateProductGoodsTimeWindowBindings(ctx context.Context, goodsID int64) error {
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT b.order_time_start, b.order_time_end
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.goods_id = ? AND b.is_deleted = 0 AND b.dock_status = 1 AND a.is_deleted = 0 AND a.status = 1
`, goodsID)
	if err != nil {
		return apiErr(consts.CodeInternalError, "库存配置校验失败")
	}

	for _, row := range rows {
		start := productGoodsRecordString(row, "order_time_start")
		end := productGoodsRecordString(row, "order_time_end")
		if _, _, normalizeErr := normalizeOrderTimeWindow(start, end); normalizeErr == nil && start != "" && end != "" {
			return nil
		}
	}
	return apiErr(consts.CodeBadRequest, "按时段提交至少需要一条配置完整时段的启用绑定")
}

func productGoodsInventoryConfigStateFromRow(row entity.ProductGoodsChannelConfig) productGoodsInventoryConfigState {
	return productGoodsInventoryConfigState{
		SmartReorderEnabled:   row.SmartReorderEnabled,
		ReorderTimeoutEnabled: row.ReorderTimeoutEnabled,
		ReorderTimeoutMinutes: row.ReorderTimeoutMinutes,
		OrderStrategy:         row.OrderStrategy,
		SyncCostPriceEnabled:  row.SyncCostPriceEnabled,
		SyncGoodsNameEnabled:  row.SyncGoodsNameEnabled,
		AllowLossSaleEnabled:  row.AllowLossSaleEnabled,
		MaxLossAmount:         row.MaxLossAmount,
		ComboGoodsEnabled:     row.ComboGoodsEnabled,
	}
}
