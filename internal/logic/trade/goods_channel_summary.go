package tradelogic

import (
	"context"
	"sort"
	"strings"
	"time"

	"myjob/internal/consts"

	"github.com/shopspring/decimal"
)

type goodsChannelBindingSummaryRow struct {
	ID          int64  `db:"id"`
	ChannelName string `db:"channel_name"`
	CostPrice   string `db:"cost_price"`
	Sort        int    `db:"sort"`
	IsAutoChange int   `db:"is_auto_change"`
	StartTime   string `db:"start_time"`
	EndTime     string `db:"end_time"`
}

func (l *TradeOrderLogic) refreshGoodsChannelSummary(ctx context.Context, goodsID int64) error {
	if goodsID <= 0 {
		return nil
	}
	now := l.core.Now()
	if err := ensureGoodsChannelConfigRow(ctx, ctxDBRunner{ctx: ctx, db: l.core.DB()}, now, goodsID); err != nil {
		return err
	}

	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT route_mode
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil || record == nil || len(record) == 0 {
		return nil
	}
	routeMode := strings.TrimSpace(record["route_mode"].String())

	bindings := make([]goodsChannelBindingSummaryRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &bindings, `
SELECT
    b.id,
    COALESCE(a.name, '') AS channel_name,
    b.cost_price,
    b.sort,
    b.is_auto_change,
    b.start_time,
    b.end_time
FROM product_goods_channel_binding b
LEFT JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.goods_id = ? AND b.is_deleted = 0 AND b.dock_status = 'enabled'
ORDER BY b.sort ASC, b.id ASC
`, goodsID); err != nil {
		return nil
	}

	boundCount := len(bindings)
	autoPrice := 0
	for _, binding := range bindings {
		if binding.IsAutoChange != 0 {
			autoPrice = 1
			break
		}
	}

	minCost := minBindingCostTrade(bindings)
	var minCostAny any = nil
	if strings.TrimSpace(minCost) != "" {
		minCostAny = minCost
	}

	var primaryBindingID any = nil
	primaryName := ""
	if len(bindings) > 0 {
		primary, name := choosePrimaryBindingTrade(routeMode, bindings, now)
		primaryName = name
		if primary != nil {
			primaryBindingID = primary.ID
		}
	}

	_, _ = l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_config
SET min_channel_cost_snapshot = ?,
    bound_channel_count_snapshot = ?,
    primary_binding_id = ?,
    primary_channel_name_snapshot = ?,
    channel_auto_price_status_snapshot = ?,
    updated_at = ?
WHERE goods_id = ?
`, minCostAny, boundCount, primaryBindingID, primaryName, autoPrice, now, goodsID)
	return nil
}

func minBindingCostTrade(bindings []goodsChannelBindingSummaryRow) string {
	var min decimal.Decimal
	hasMin := false
	for _, binding := range bindings {
		value, ok := parseMoneyDecimalTrade(binding.CostPrice)
		if !ok {
			continue
		}
		if !hasMin || value.LessThan(min) {
			min = value
			hasMin = true
		}
	}
	if !hasMin {
		return ""
	}
	return Round4(min).StringFixed(4)
}

func choosePrimaryBindingTrade(routeMode string, bindings []goodsChannelBindingSummaryRow, now time.Time) (*goodsChannelBindingSummaryRow, string) {
	switch strings.TrimSpace(routeMode) {
	case RouteModeWeightPercent, RouteModeRandom:
		return nil, "按规则选路"
	case RouteModeLowestCostFirst:
		candidates := append([]goodsChannelBindingSummaryRow(nil), bindings...)
		sort.Slice(candidates, func(i, j int) bool {
			left, leftOK := parseMoneyDecimalTrade(candidates[i].CostPrice)
			right, rightOK := parseMoneyDecimalTrade(candidates[j].CostPrice)
			if leftOK != rightOK {
				return leftOK
			}
			if leftOK && rightOK && !left.Equal(right) {
				return left.LessThan(right)
			}
			if candidates[i].Sort != candidates[j].Sort {
				return candidates[i].Sort < candidates[j].Sort
			}
			return candidates[i].ID < candidates[j].ID
		})
		primary := candidates[0]
		return &primary, strings.TrimSpace(primary.ChannelName)
	case RouteModeTimePeriod:
		matched := make([]goodsChannelBindingSummaryRow, 0)
		for _, binding := range bindings {
			if inTimePeriod(now, binding.StartTime, binding.EndTime) {
				matched = append(matched, binding)
			}
		}
		if len(matched) > 0 {
			primary := matched[0]
			return &primary, strings.TrimSpace(primary.ChannelName)
		}
		fallthrough
	default:
		primary := bindings[0]
		return &primary, strings.TrimSpace(primary.ChannelName)
	}
}

func parseMoneyDecimalTrade(value string) (decimal.Decimal, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, false
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, false
	}
	return amount, true
}

var _ = consts.CodeInternalError

