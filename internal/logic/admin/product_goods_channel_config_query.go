package adminlogic

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/consts"

	"github.com/shopspring/decimal"
)

type goodsChannelConfigGoodsRow struct {
	ID               int64          `db:"id"`
	GoodsCode        string         `db:"goods_code"`
	GoodsName        string         `db:"goods_name"`
	SupplyType       string         `db:"supply_type"`
	SubjectID        sql.NullInt64  `db:"subject_id"`
	SubjectName      string         `db:"subject_name"`
	DefaultSellPrice sql.NullString `db:"default_sell_price"`
}

type goodsChannelConfigRow struct {
	GoodsID                        int64          `db:"goods_id"`
	SmartReplenishEnabled          int            `db:"smart_replenish_enabled"`
	AttemptTimeoutEnabled          int            `db:"attempt_timeout_enabled"`
	AttemptTimeoutMinutes          int            `db:"attempt_timeout_minutes"`
	RouteMode                      string         `db:"route_mode"`
	SyncCostEnabled                int            `db:"sync_cost_enabled"`
	SyncGoodsNameEnabled           int            `db:"sync_goods_name_enabled"`
	AllowLoss                      int            `db:"allow_loss"`
	MaxLossAmount                  sql.NullString `db:"max_loss_amount"`
	IsBundle                       int            `db:"is_bundle"`
	MinChannelCostSnapshot         sql.NullString `db:"min_channel_cost_snapshot"`
	BoundChannelCountSnapshot      int            `db:"bound_channel_count_snapshot"`
	PrimaryBindingID               sql.NullInt64  `db:"primary_binding_id"`
	PrimaryChannelNameSnapshot     string         `db:"primary_channel_name_snapshot"`
	ChannelAutoPriceStatusSnapshot int            `db:"channel_auto_price_status_snapshot"`
}

type goodsChannelBindingRow struct {
	ID           int64  `db:"id"`
	ChannelName  string `db:"channel_name"`
	CostPrice    string `db:"cost_price"`
	Sort         int    `db:"sort"`
	IsAutoChange int    `db:"is_auto_change"`
	StartTime    string `db:"start_time"`
	EndTime      string `db:"end_time"`
}

// Get 返回指定商品的渠道配置与绑定摘要（若不存在配置行则自动初始化默认值）。
func (l *ProductGoodsChannelConfigLogic) Get(ctx context.Context, req *adminapi.ProductGoodsChannelConfigGetReq) (*adminapi.ProductGoodsChannelConfigGetRes, error) {
	goods, err := l.getActiveGoodsForChannelConfig(ctx, req.GoodsID)
	if err != nil {
		return nil, err
	}
	if goods.SupplyType != "channel" {
		return nil, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}
	if err := l.ensureGoodsChannelConfigRow(ctx, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取商品渠道配置失败")
	}
	if err := l.refreshGoodsChannelSummary(ctx, req.GoodsID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取商品渠道配置失败")
	}
	cfg, err := l.getGoodsChannelConfigRow(ctx, req.GoodsID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取商品渠道配置失败")
	}

	var subjectID *int64
	if goods.SubjectID.Valid {
		id := goods.SubjectID.Int64
		subjectID = &id
	}

	var maxLossAmount *string
	if cfg.MaxLossAmount.Valid {
		value := strings.TrimSpace(cfg.MaxLossAmount.String)
		if value != "" {
			normalized := formatMoney(value)
			maxLossAmount = &normalized
		}
	}

	minChannelCost := ""
	if cfg.MinChannelCostSnapshot.Valid {
		minChannelCost = formatMoney(cfg.MinChannelCostSnapshot.String)
	}

	return &adminapi.ProductGoodsChannelConfigGetRes{
		GoodsID:          goods.ID,
		GoodsCode:        goods.GoodsCode,
		GoodsName:        goods.GoodsName,
		SubjectID:        subjectID,
		SubjectName:      goods.SubjectName,
		DefaultSellPrice: formatMoney(goods.DefaultSellPrice.String),

		SmartReplenishEnabled: boolFlag(cfg.SmartReplenishEnabled),
		AttemptTimeoutEnabled: boolFlag(cfg.AttemptTimeoutEnabled),
		AttemptTimeoutMinutes: cfg.AttemptTimeoutMinutes,
		RouteMode:             cfg.RouteMode,
		SyncCostEnabled:       boolFlag(cfg.SyncCostEnabled),
		SyncGoodsNameEnabled:  boolFlag(cfg.SyncGoodsNameEnabled),
		AllowLoss:             boolFlag(cfg.AllowLoss),
		MaxLossAmount:         maxLossAmount,
		IsBundle:              boolFlag(cfg.IsBundle),

		BoundChannelCount:      cfg.BoundChannelCountSnapshot,
		PrimaryChannelName:     cfg.PrimaryChannelNameSnapshot,
		MinChannelCost:         minChannelCost,
		ChannelAutoPriceStatus: boolFlag(cfg.ChannelAutoPriceStatusSnapshot),
	}, nil
}

func (l *ProductGoodsChannelConfigLogic) getActiveGoodsForChannelConfig(ctx context.Context, goodsID int64) (goodsChannelConfigGoodsRow, error) {
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    p.id,
    p.goods_code,
    p.name AS goods_name,
    p.supply_type,
    p.subject_id,
    COALESCE(sub.name, '') AS subject_name,
    p.default_sell_price
FROM product_goods p
LEFT JOIN admin_subject sub ON sub.id = p.subject_id
WHERE p.id = ? AND p.is_deleted = 0
`, goodsID)
	if err != nil {
		return goodsChannelConfigGoodsRow{}, apiErr(consts.CodeInternalError, "读取商品失败")
	}
	if record == nil || len(record) == 0 {
		return goodsChannelConfigGoodsRow{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	return goodsChannelConfigGoodsRow{
		ID:               record["id"].Int64(),
		GoodsCode:        record["goods_code"].String(),
		GoodsName:        record["goods_name"].String(),
		SupplyType:       record["supply_type"].String(),
		SubjectID:        productGoodsRecordNullInt64(record, "subject_id"),
		SubjectName:      productGoodsRecordString(record, "subject_name"),
		DefaultSellPrice: productGoodsRecordNullString(record, "default_sell_price"),
	}, nil
}

func (l *ProductGoodsChannelConfigLogic) ensureGoodsChannelConfigRow(ctx context.Context, goodsID int64) error {
	exists, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods_channel_config WHERE goods_id = ?`, goodsID)
	if err != nil {
		return err
	}
	if exists.Int() > 0 {
		return nil
	}
	_, err = l.core.DB().Exec(ctx, `INSERT INTO product_goods_channel_config (goods_id, created_at, updated_at) VALUES (?, ?, ?)`, goodsID, l.core.Now(), l.core.Now())
	return err
}

func (l *ProductGoodsChannelConfigLogic) getGoodsChannelConfigRow(ctx context.Context, goodsID int64) (goodsChannelConfigRow, error) {
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    goods_id,
    smart_replenish_enabled,
    attempt_timeout_enabled,
    attempt_timeout_minutes,
    route_mode,
    sync_cost_enabled,
    sync_goods_name_enabled,
    allow_loss,
    max_loss_amount,
    is_bundle,
    min_channel_cost_snapshot,
    bound_channel_count_snapshot,
    primary_binding_id,
    primary_channel_name_snapshot,
    channel_auto_price_status_snapshot
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return goodsChannelConfigRow{}, err
	}
	if record == nil || len(record) == 0 {
		return goodsChannelConfigRow{}, apiErr(consts.CodeBadRequest, "商品渠道配置不存在")
	}
	return goodsChannelConfigRow{
		GoodsID:                        record["goods_id"].Int64(),
		SmartReplenishEnabled:          record["smart_replenish_enabled"].Int(),
		AttemptTimeoutEnabled:          record["attempt_timeout_enabled"].Int(),
		AttemptTimeoutMinutes:          record["attempt_timeout_minutes"].Int(),
		RouteMode:                      record["route_mode"].String(),
		SyncCostEnabled:                record["sync_cost_enabled"].Int(),
		SyncGoodsNameEnabled:           record["sync_goods_name_enabled"].Int(),
		AllowLoss:                      record["allow_loss"].Int(),
		MaxLossAmount:                  productGoodsRecordNullString(record, "max_loss_amount"),
		IsBundle:                       record["is_bundle"].Int(),
		MinChannelCostSnapshot:         productGoodsRecordNullString(record, "min_channel_cost_snapshot"),
		BoundChannelCountSnapshot:      record["bound_channel_count_snapshot"].Int(),
		PrimaryBindingID:               productGoodsRecordNullInt64(record, "primary_binding_id"),
		PrimaryChannelNameSnapshot:     record["primary_channel_name_snapshot"].String(),
		ChannelAutoPriceStatusSnapshot: record["channel_auto_price_status_snapshot"].Int(),
	}, nil
}

func boolFlag(value int) bool { return value != 0 }

func (l *ProductGoodsChannelConfigLogic) refreshGoodsChannelSummary(ctx context.Context, goodsID int64) error {
	if err := l.ensureGoodsChannelConfigRow(ctx, goodsID); err != nil {
		return err
	}
	cfg, err := l.getGoodsChannelConfigRow(ctx, goodsID)
	if err != nil {
		return err
	}
	bindings, err := l.loadEnabledGoodsChannelBindings(ctx, goodsID)
	if err != nil {
		return err
	}

	boundCount := len(bindings)
	autoPrice := 0
	minCost := ""
	if len(bindings) > 0 {
		minCost = minBindingCost(bindings)
		for _, binding := range bindings {
			if binding.IsAutoChange != 0 {
				autoPrice = 1
				break
			}
		}
	}

	var primaryBindingID any = nil
	primaryName := ""
	if len(bindings) > 0 {
		primary, name := choosePrimaryBinding(cfg.RouteMode, bindings, l.core.Now())
		primaryName = name
		if primary != nil {
			primaryBindingID = primary.ID
		}
	}

	var minCostAny any = nil
	if strings.TrimSpace(minCost) != "" {
		minCostAny = minCost
	}

	_, err = l.core.DB().Exec(ctx, `
UPDATE product_goods_channel_config
SET min_channel_cost_snapshot = ?,
    bound_channel_count_snapshot = ?,
    primary_binding_id = ?,
    primary_channel_name_snapshot = ?,
    channel_auto_price_status_snapshot = ?,
    updated_at = ?
WHERE goods_id = ?
`, minCostAny, boundCount, primaryBindingID, primaryName, autoPrice, l.core.Now(), goodsID)
	return err
}

func (l *ProductGoodsChannelConfigLogic) loadEnabledGoodsChannelBindings(ctx context.Context, goodsID int64) ([]goodsChannelBindingRow, error) {
	rows := make([]goodsChannelBindingRow, 0)
	err := l.core.DB().GetCore().GetScan(ctx, &rows, `
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
`, goodsID)
	return rows, err
}

func minBindingCost(bindings []goodsChannelBindingRow) string {
	var min decimal.Decimal
	hasMin := false
	for _, binding := range bindings {
		value, ok := parseMoneyDecimal(binding.CostPrice)
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
	return min.StringFixed(4)
}

func choosePrimaryBinding(routeMode string, bindings []goodsChannelBindingRow, now time.Time) (*goodsChannelBindingRow, string) {
	switch strings.TrimSpace(routeMode) {
	case "weight_percent", "random":
		return nil, "按规则选路"
	case "lowest_cost_first":
		candidates := append([]goodsChannelBindingRow(nil), bindings...)
		sort.Slice(candidates, func(i, j int) bool {
			left, leftOK := parseMoneyDecimal(candidates[i].CostPrice)
			right, rightOK := parseMoneyDecimal(candidates[j].CostPrice)
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
		return &primary, primary.ChannelName
	case "time_period":
		matched := make([]goodsChannelBindingRow, 0)
		for _, binding := range bindings {
			if inTimePeriod(now, binding.StartTime, binding.EndTime) {
				matched = append(matched, binding)
			}
		}
		if len(matched) > 0 {
			primary := matched[0]
			return &primary, primary.ChannelName
		}
		fallthrough
	default:
		primary := bindings[0]
		return &primary, primary.ChannelName
	}
}

func parseMoneyDecimal(value string) (decimal.Decimal, bool) {
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

func inTimePeriod(now time.Time, startTime string, endTime string) bool {
	startMin, ok := parseTimeHM(startTime)
	if !ok {
		return false
	}
	endMin, ok := parseTimeHM(endTime)
	if !ok {
		return false
	}
	nowMin := now.Hour()*60 + now.Minute()
	if startMin == endMin {
		return false
	}
	if startMin < endMin {
		return nowMin >= startMin && nowMin < endMin
	}
	// 跨天时段，例如 23:00-02:00。
	return nowMin >= startMin || nowMin < endMin
}

func parseTimeHM(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if len(value) != 5 || value[2] != ':' {
		return 0, false
	}
	hour := int(value[0]-'0')*10 + int(value[1]-'0')
	min := int(value[3]-'0')*10 + int(value[4]-'0')
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return 0, false
	}
	return hour*60 + min, true
}
