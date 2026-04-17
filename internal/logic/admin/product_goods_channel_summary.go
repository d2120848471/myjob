package adminlogic

import (
	"context"
	"strings"
)

type goodsChannelSummary struct {
	BoundChannels          []string
	BoundChannelCount      int
	PrimaryChannelName     string
	MinChannelCost         string
	ChannelAutoPriceStatus bool
}

type goodsChannelBindingListRow struct {
	GoodsID      int64  `db:"goods_id"`
	ID           int64  `db:"id"`
	ChannelName  string `db:"channel_name"`
	CostPrice    string `db:"cost_price"`
	Sort         int    `db:"sort"`
	IsAutoChange int    `db:"is_auto_change"`
	StartTime    string `db:"start_time"`
	EndTime      string `db:"end_time"`
}

func (l *ProductGoodsLogic) loadGoodsChannelSummaries(ctx context.Context, goodsIDs []int64) (map[int64]goodsChannelSummary, error) {
	summaries := make(map[int64]goodsChannelSummary, len(goodsIDs))
	if len(goodsIDs) == 0 {
		return summaries, nil
	}

	args := make([]any, 0, len(goodsIDs))
	for _, goodsID := range goodsIDs {
		args = append(args, goodsID)
	}

	type routeRow struct {
		GoodsID   int64  `db:"goods_id"`
		RouteMode string `db:"route_mode"`
	}
	routeRows := make([]routeRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &routeRows, `
SELECT goods_id, route_mode
FROM product_goods_channel_config
WHERE goods_id IN (`+sqlPlaceholders(len(goodsIDs))+`)
`, args...); err != nil {
		return nil, err
	}
	routeModeByGoodsID := make(map[int64]string, len(routeRows))
	for _, row := range routeRows {
		routeModeByGoodsID[row.GoodsID] = row.RouteMode
	}

	bindingRows := make([]goodsChannelBindingListRow, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &bindingRows, `
SELECT
    b.goods_id,
    b.id,
    COALESCE(a.name, '') AS channel_name,
    b.cost_price,
    b.sort,
    b.is_auto_change,
    b.start_time,
    b.end_time
FROM product_goods_channel_binding b
LEFT JOIN supplier_platform_account a ON a.id = b.platform_account_id
WHERE b.goods_id IN (`+sqlPlaceholders(len(goodsIDs))+`) AND b.is_deleted = 0 AND b.dock_status = 'enabled'
ORDER BY b.goods_id ASC, b.sort ASC, b.id ASC
`, args...); err != nil {
		return nil, err
	}

	bindingsByGoodsID := make(map[int64][]goodsChannelBindingRow, len(goodsIDs))
	for _, row := range bindingRows {
		bindingsByGoodsID[row.GoodsID] = append(bindingsByGoodsID[row.GoodsID], goodsChannelBindingRow{
			ID:           row.ID,
			ChannelName:  row.ChannelName,
			CostPrice:    row.CostPrice,
			Sort:         row.Sort,
			IsAutoChange: row.IsAutoChange,
			StartTime:    row.StartTime,
			EndTime:      row.EndTime,
		})
	}

	now := l.core.Now()
	for _, goodsID := range goodsIDs {
		routeMode := strings.TrimSpace(routeModeByGoodsID[goodsID])
		bindings := bindingsByGoodsID[goodsID]
		summary := goodsChannelSummary{
			BoundChannels:          []string{},
			BoundChannelCount:      len(bindings),
			PrimaryChannelName:     "",
			MinChannelCost:         "",
			ChannelAutoPriceStatus: false,
		}
		if len(bindings) > 0 {
			summary.MinChannelCost = minBindingCost(bindings)
			for _, binding := range bindings {
				if binding.IsAutoChange != 0 {
					summary.ChannelAutoPriceStatus = true
					break
				}
			}
			_, primaryName := choosePrimaryBinding(routeMode, bindings, now)
			summary.PrimaryChannelName = primaryName

			seen := make(map[string]struct{}, len(bindings))
			for _, binding := range bindings {
				name := strings.TrimSpace(binding.ChannelName)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				summary.BoundChannels = append(summary.BoundChannels, name)
			}
		}
		summaries[goodsID] = summary
	}

	return summaries, nil
}
