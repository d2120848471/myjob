package adminlogic

import (
	"context"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

const (
	purchaseLimitTypeMember  = 1
	purchaseLimitTypeAccount = 2

	purchaseLimitPeriodDay      = 1
	purchaseLimitPeriodInterval = 2
)

var purchaseLimitTypeItems = []entity.PurchaseLimitEnumItem{
	{ID: purchaseLimitTypeMember, Title: "同一会员"},
	{ID: purchaseLimitTypeAccount, Title: "同一充值账号"},
}

var purchaseLimitPeriodTypeItems = []entity.PurchaseLimitEnumItem{
	{ID: purchaseLimitPeriodDay, Title: "按天"},
	{ID: purchaseLimitPeriodInterval, Title: "按区间(分钟)"},
}

var purchaseLimitTypeTitles = enumTitleMap(purchaseLimitTypeItems)
var purchaseLimitPeriodTypeTitles = enumTitleMap(purchaseLimitPeriodTypeItems)

type PurchaseLimitLogic struct{ core *app.Core }

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

func (l *PurchaseLimitLogic) Add(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyCreateRes, error) {
	normalized, err := normalizePurchaseLimitInput(req.Name, req.LimitType, req.PeriodType, req.Period, req.LimitNums, req.LimitTimes)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	// 新增弹层不承载状态切换，后端默认开通，后续启停统一走独立状态接口。
	defaultStatus := consts.StatusEnabled
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_purchase_limit_strategy (
    name, limit_type, period_type, period, limit_nums, limit_times, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, normalized.Name, normalized.LimitType, normalized.PeriodType, normalized.Period, normalized.LimitNums, normalized.LimitTimes, defaultStatus, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品购买数量限制策略：%s", normalized.Name), ip)
	return &adminapi.PurchaseLimitStrategyCreateRes{ID: id}, nil
}

func (l *PurchaseLimitLogic) Edit(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyUpdateRes, error) {
	if _, err := l.getStrategy(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	normalized, err := normalizePurchaseLimitInput(req.Name, req.LimitType, req.PeriodType, req.Period, req.LimitNums, req.LimitTimes)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	// 编辑弹层只更新策略内容，不顺带改状态，避免未点启停却把策略误关停。
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_purchase_limit_strategy
SET name = ?, limit_type = ?, period_type = ?, period = ?, limit_nums = ?, limit_times = ?, updated_at = ?
WHERE id = ?
`, normalized.Name, normalized.LimitType, normalized.PeriodType, normalized.Period, normalized.LimitNums, normalized.LimitTimes, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品购买数量限制策略：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.PurchaseLimitStrategyUpdateRes{}, nil
}

func (l *PurchaseLimitLogic) Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyDeleteRes, error) {
	strategy, err := l.getStrategy(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_purchase_limit_strategy WHERE id = ?`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品购买数量限制策略：%d -> %s", req.ID, strategy.Name), ip)
	return &adminapi.PurchaseLimitStrategyDeleteRes{}, nil
}

func (l *PurchaseLimitLogic) Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyStatusRes, error) {
	strategy, err := l.getStrategy(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品购买数量限制策略不存在")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_purchase_limit_strategy
SET status = ?, updated_at = ?
WHERE id = ?
`, req.Status, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品购买数量限制策略状态更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换商品购买数量限制策略状态：%d -> %s -> %d", req.ID, strategy.Name, req.Status), ip)
	return &adminapi.PurchaseLimitStrategyStatusRes{}, nil
}

func (l *PurchaseLimitLogic) Enums(ctx context.Context, _ *adminapi.PurchaseLimitStrategyEnumsReq) (*adminapi.PurchaseLimitStrategyEnumsRes, error) {
	limitTypes := append([]entity.PurchaseLimitEnumItem(nil), purchaseLimitTypeItems...)
	periodTypes := append([]entity.PurchaseLimitEnumItem(nil), purchaseLimitPeriodTypeItems...)
	return &adminapi.PurchaseLimitStrategyEnumsRes{
		LimitTypes:  limitTypes,
		PeriodTypes: periodTypes,
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

type normalizedPurchaseLimitInput struct {
	Name       string
	LimitType  int
	PeriodType int
	Period     int
	LimitNums  int
	LimitTimes int
}

func normalizePurchaseLimitInput(name string, limitType, periodType, period, limitNums, limitTimes int) (normalizedPurchaseLimitInput, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("策略名称不能为空")
	}
	if _, ok := purchaseLimitTypeTitles[limitType]; !ok {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制类型错误")
	}
	if _, ok := purchaseLimitPeriodTypeTitles[periodType]; !ok {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("周期类型错误")
	}
	if period <= 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制周期必须大于0")
	}
	// 数量和笔数允许填 0，语义是“不限制”，但不允许负数落库。
	if limitNums < 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制数量不能小于0")
	}
	if limitTimes < 0 {
		return normalizedPurchaseLimitInput{}, fmt.Errorf("限制笔数不能小于0")
	}
	return normalizedPurchaseLimitInput{
		Name:       name,
		LimitType:  limitType,
		PeriodType: periodType,
		Period:     period,
		LimitNums:  limitNums,
		LimitTimes: limitTimes,
	}, nil
}

func enumTitleMap(items []entity.PurchaseLimitEnumItem) map[int]string {
	result := make(map[int]string, len(items))
	for _, item := range items {
		result[item.ID] = item.Title
	}
	return result
}
