package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// ListRules 分页查询充值账号风控规则。
func (l *RechargeRiskLogic) ListRules(ctx context.Context, req *adminapi.RechargeRiskRuleListReq) (*adminapi.RechargeRiskRuleListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"is_deleted = 0"}
	args := make([]any, 0, 8)
	if account := strings.TrimSpace(req.Account); account != "" {
		conditions = append(conditions, "account LIKE ?")
		args = append(args, "%"+account+"%")
	}
	if keyword := strings.TrimSpace(req.GoodsKeyword); keyword != "" {
		conditions = append(conditions, "goods_keyword LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if status, ok, err := normalizeRechargeRiskStatusFilter(req.Status); err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	} else if ok {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_rule WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则列表查询失败")
	}
	rows := make([]entity.RechargeRiskRule, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, account, goods_keyword, reason, status, hit_count, created_by_id, created_by_name,
       updated_by_id, updated_by_name, is_deleted, deleted_at, created_at, updated_at
FROM recharge_risk_rule
WHERE `+whereClause+`
ORDER BY updated_at DESC, id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控规则列表查询失败")
	}
	items := make([]adminapi.RechargeRiskRuleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rechargeRiskRuleItemFromEntity(row))
	}
	return &adminapi.RechargeRiskRuleListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

// ListRecords 分页查询充值账号风控拦截记录。
func (l *RechargeRiskLogic) ListRecords(ctx context.Context, req *adminapi.RechargeRiskRecordListReq) (*adminapi.RechargeRiskRecordListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 8)
	if account := strings.TrimSpace(req.Account); account != "" {
		conditions = append(conditions, "account LIKE ?")
		args = append(args, "%"+account+"%")
	}
	if keyword := strings.TrimSpace(req.GoodsKeyword); keyword != "" {
		conditions = append(conditions, "matched_keyword LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if startTime := strings.TrimSpace(req.StartTime); startTime != "" {
		parsed, err := app.ParseQueryTime(startTime)
		if err != nil {
			return nil, apiErr(consts.CodeBadRequest, "拦截开始时间格式错误")
		}
		conditions = append(conditions, "intercepted_at >= ?")
		args = append(args, parsed)
	}
	if endTime := strings.TrimSpace(req.EndTime); endTime != "" {
		parsed, err := app.ParseQueryTime(endTime)
		if err != nil {
			return nil, apiErr(consts.CodeBadRequest, "拦截结束时间格式错误")
		}
		conditions = append(conditions, "intercepted_at <= ?")
		args = append(args, parsed)
	}
	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM recharge_risk_record WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控记录列表查询失败")
	}
	rows := make([]entity.RechargeRiskRecord, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id, rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
       matched_keyword, reason, request_token_masked, intercepted_at, created_at
FROM recharge_risk_record
WHERE `+whereClause+`
ORDER BY intercepted_at DESC, id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "风控记录列表查询失败")
	}
	items := make([]adminapi.RechargeRiskRecordItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rechargeRiskRecordItemFromEntity(row))
	}
	return &adminapi.RechargeRiskRecordListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}
