package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// List 分页查询购买数量限制策略列表，支持按名称关键字筛选。
func (l *PurchaseLimitLogic) List(ctx context.Context, req *adminapi.PurchaseLimitStrategyListReq) (*adminapi.PurchaseLimitStrategyListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	keyword := strings.TrimSpace(req.Keyword)

	// 当前列表页只支持按策略名称做关键词过滤，和页面查询能力保持一致。
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 4)
	if keyword != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+keyword+"%")
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_purchase_limit_strategy WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略列表查询失败")
	}

	rows := make([]entity.PurchaseLimitStrategy, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    id,
    name,
    limit_type,
    period_type,
    period,
    limit_nums,
    limit_times,
    status,
    created_at,
    updated_at
FROM product_purchase_limit_strategy
WHERE `+whereClause+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略列表查询失败")
	}

	items := make([]entity.PurchaseLimitStrategyListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.PurchaseLimitStrategyListItem{
			ID:              row.ID,
			Name:            row.Name,
			LimitType:       row.LimitType,
			LimitTypeLabel:  purchaseLimitTypeTitles[row.LimitType],
			PeriodType:      row.PeriodType,
			PeriodTypeLabel: purchaseLimitPeriodTypeTitles[row.PeriodType],
			Period:          row.Period,
			LimitNums:       row.LimitNums,
			LimitTimes:      row.LimitTimes,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return &adminapi.PurchaseLimitStrategyListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

func (l *PurchaseLimitLogic) getStrategy(ctx context.Context, id int64) (entity.PurchaseLimitStrategy, error) {
	strategy := entity.PurchaseLimitStrategy{}
	err := l.core.DB().GetCore().GetScan(ctx, &strategy, `
SELECT
    id,
    name,
    limit_type,
    period_type,
    period,
    limit_nums,
    limit_times,
    status,
    created_at,
    updated_at
FROM product_purchase_limit_strategy
WHERE id = ?
`, id)
	return strategy, err
}
