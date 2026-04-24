package orderlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// QueryOpenOrder 返回外部可见的订单状态，不暴露渠道和成本。
func (l *OrderLogic) QueryOpenOrder(ctx context.Context, token, orderNo string) (*adminapi.OpenOrderQueryRes, error) {
	if strings.TrimSpace(token) != strings.TrimSpace(l.core.Config().OpenOrder.Token) {
		return nil, unauthorizedErr()
	}
	order, err := l.getOrderByNo(ctx, strings.TrimSpace(orderNo))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "订单不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "订单查询失败")
	}
	return &adminapi.OpenOrderQueryRes{
		OrderNo:    order.OrderNo,
		StatusCode: order.Status,
		StatusText: orderStatusText(order.Status),
		GoodsID:    order.GoodsCode,
		GoodsName:  order.GoodsName,
		Account:    order.Account,
		Quantity:   order.Quantity,
		CreatedAt:  formatAppTime(order.CreatedAt),
		UpdatedAt:  formatAppTime(order.UpdatedAt),
	}, nil
}

type adminOrderFilter struct {
	Conditions []string
	Args       []any
}

// ListAdminOrders 分页查询后台订单列表，并返回同一筛选条件下的今日/昨日统计。
func (l *OrderLogic) ListAdminOrders(ctx context.Context, req *adminapi.OrderListReq) (*adminapi.OrderListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	filter, err := l.buildAdminOrderFilter(req)
	if err != nil {
		return nil, err
	}
	whereClause := strings.Join(filter.Conditions, " AND ")

	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM external_order o
LEFT JOIN external_order_attempt a ON a.id = o.current_attempt_id
WHERE `+whereClause, filter.Args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "订单列表查询失败")
	}

	queryArgs := append(append([]any{}, filter.Args...), pageSize, (page-1)*pageSize)
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    o.id,
    o.subject_name,
    o.order_no,
    o.goods_code,
    o.goods_name,
    o.account,
    o.quantity,
    o.order_amount,
    o.cost_amount,
    o.profit_amount,
    o.attempt_count,
    o.last_receipt,
    o.status,
    o.created_at,
    o.updated_at,
    COALESCE(a.platform_account_id, 0) AS current_channel_id,
    COALESCE(a.platform_account_name, '') AS current_channel_name,
    COALESCE(a.supplier_order_no, '') AS supplier_order_no
FROM external_order o
LEFT JOIN external_order_attempt a ON a.id = o.current_attempt_id
WHERE `+whereClause+`
ORDER BY o.created_at DESC, o.id DESC
LIMIT ? OFFSET ?
`, queryArgs...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "订单列表查询失败")
	}

	items := make([]adminapi.OrderListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, adminOrderListItemFromRecord(row))
	}
	stats, err := l.loadAdminOrderStats(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &adminapi.OrderListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
		Stats:      stats,
	}, nil
}

func (l *OrderLogic) getOrderByNo(ctx context.Context, orderNo string) (entity.ExternalOrder, error) {
	row := entity.ExternalOrder{}
	err := l.core.DB().GetCore().GetScan(ctx, &row, `SELECT * FROM external_order WHERE order_no = ?`, orderNo)
	return row, err
}

func (l *OrderLogic) buildAdminOrderFilter(req *adminapi.OrderListReq) (adminOrderFilter, error) {
	conditions := []string{"1=1"}
	args := make([]any, 0, 12)
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		switch strings.TrimSpace(strings.ToLower(req.KeywordBy)) {
		case "order_no":
			conditions = append(conditions, "o.order_no LIKE ?")
			args = append(args, likeKeyword)
		case "account":
			conditions = append(conditions, "o.account LIKE ?")
			args = append(args, likeKeyword)
		case "goods_name":
			conditions = append(conditions, "o.goods_name LIKE ?")
			args = append(args, likeKeyword)
		default:
			conditions = append(conditions, "(o.order_no LIKE ? OR o.account LIKE ? OR o.goods_name LIKE ?)")
			args = append(args, likeKeyword, likeKeyword, likeKeyword)
		}
	}
	if status := strings.TrimSpace(req.Status); status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, status)
	}
	if hasTax, hasFilter, err := normalizeAdminOrderBinaryFilter(req.HasTax, "含税状态"); err != nil {
		return adminOrderFilter{}, err
	} else if hasFilter {
		conditions = append(conditions, "o.has_tax = ?")
		args = append(args, hasTax)
	}
	if req.ChannelID > 0 {
		conditions = append(conditions, "a.platform_account_id = ?")
		args = append(args, req.ChannelID)
	}
	if isCard, hasFilter, err := normalizeAdminOrderBinaryFilter(req.IsCard, "卡密状态"); err != nil {
		return adminOrderFilter{}, err
	} else if hasFilter && isCard == 1 {
		// 当前开放订单只覆盖直充渠道商品；卡密筛选先返回空结果，避免误把直充订单混入卡密视图。
		conditions = append(conditions, "1=0")
	}
	if err := l.appendAdminOrderTimeFilters(req, &conditions, &args); err != nil {
		return adminOrderFilter{}, err
	}
	return adminOrderFilter{Conditions: conditions, Args: args}, nil
}

func normalizeAdminOrderBinaryFilter(value, label string) (int, bool, error) {
	switch strings.TrimSpace(value) {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, apiErr(consts.CodeBadRequest, label+"筛选值错误")
	}
}

func (l *OrderLogic) appendAdminOrderTimeFilters(req *adminapi.OrderListReq, conditions *[]string, args *[]any) error {
	if start, end, ok, err := adminOrderQuickRange(strings.TrimSpace(req.QuickRange), l.core.Now()); err != nil {
		return apiErr(consts.CodeBadRequest, "快捷时间范围错误")
	} else if ok {
		*conditions = append(*conditions, "o.created_at >= ?", "o.created_at < ?")
		*args = append(*args, start, end)
	}
	if strings.TrimSpace(req.StartTime) != "" {
		parsed, err := app.ParseQueryTime(req.StartTime)
		if err != nil {
			return apiErr(consts.CodeBadRequest, "时间范围格式错误")
		}
		*conditions = append(*conditions, "o.created_at >= ?")
		*args = append(*args, parsed)
	}
	if strings.TrimSpace(req.EndTime) != "" {
		parsed, err := app.ParseQueryTime(req.EndTime)
		if err != nil {
			return apiErr(consts.CodeBadRequest, "时间范围格式错误")
		}
		*conditions = append(*conditions, "o.created_at <= ?")
		*args = append(*args, parsed)
	}
	return nil
}

func adminOrderQuickRange(value string, now time.Time) (time.Time, time.Time, bool, error) {
	if value == "" {
		return time.Time{}, time.Time{}, false, nil
	}
	todayStart := startOfDay(now)
	switch value {
	case "today":
		return todayStart, todayStart.AddDate(0, 0, 1), true, nil
	case "yesterday":
		return todayStart.AddDate(0, 0, -1), todayStart, true, nil
	case "week":
		return todayStart.AddDate(0, 0, -6), todayStart.AddDate(0, 0, 1), true, nil
	case "month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), todayStart.AddDate(0, 0, 1), true, nil
	case "three_months":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -2, 0), todayStart.AddDate(0, 0, 1), true, nil
	default:
		return time.Time{}, time.Time{}, false, errors.New("invalid quick range")
	}
}

func (l *OrderLogic) loadAdminOrderStats(ctx context.Context, filter adminOrderFilter) (adminapi.OrderStats, error) {
	todayStart := startOfDay(l.core.Now())
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	todayCount, todayAmount, err := l.sumAdminOrders(ctx, filter, todayStart, tomorrowStart)
	if err != nil {
		return adminapi.OrderStats{}, err
	}
	yesterdayCount, yesterdayAmount, err := l.sumAdminOrders(ctx, filter, yesterdayStart, todayStart)
	if err != nil {
		return adminapi.OrderStats{}, err
	}
	return adminapi.OrderStats{
		TodayOrderCount:      todayCount,
		TodayOrderAmount:     todayAmount,
		YesterdayOrderCount:  yesterdayCount,
		YesterdayOrderAmount: yesterdayAmount,
	}, nil
}

func (l *OrderLogic) sumAdminOrders(ctx context.Context, filter adminOrderFilter, start, end time.Time) (int, string, error) {
	conditions := append([]string{}, filter.Conditions...)
	args := append([]any{}, filter.Args...)
	conditions = append(conditions, "o.created_at >= ?", "o.created_at < ?")
	args = append(args, start, end)
	row, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT COUNT(*) AS order_count, COALESCE(SUM(o.order_amount), 0) AS order_amount
FROM external_order o
LEFT JOIN external_order_attempt a ON a.id = o.current_attempt_id
WHERE `+strings.Join(conditions, " AND "), args...)
	if err != nil {
		return 0, "", apiErr(consts.CodeInternalError, "订单统计查询失败")
	}
	return row["order_count"].Int(), formatOrderMoney(row["order_amount"].String()), nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
