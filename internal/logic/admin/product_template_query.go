package adminlogic

import (
	"context"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// List 分页查询商品模板列表，支持按关键字/类型/共享状态筛选。
func (l *ProductTemplateLogic) List(ctx context.Context, req *adminapi.ProductTemplateListReq) (*adminapi.ProductTemplateListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	keyword := strings.TrimSpace(req.Keyword)
	templateType, err := normalizeProductTemplateType(req.Type, false)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	isShared, hasSharedFilter, err := normalizeProductTemplateSharedFilter(req.IsShared)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	conditions := []string{"1 = 1"}
	args := make([]any, 0, 8)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		conditions = append(conditions, "(title LIKE ? OR account_name LIKE ?)")
		args = append(args, likeKeyword, likeKeyword)
	}
	if templateType != "" {
		conditions = append(conditions, "template_type = ?")
		args = append(args, templateType)
	}
	if hasSharedFilter {
		conditions = append(conditions, "is_shared = ?")
		args = append(args, isShared)
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_template WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板列表查询失败")
	}

	rows := make([]entity.ProductTemplate, 0)
	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	if err = l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    id,
    title,
    template_type,
    is_shared,
    account_name,
    validate_type,
    created_at,
    updated_at
FROM product_template
WHERE `+whereClause+`
ORDER BY id DESC
LIMIT ? OFFSET ?
`, queryArgs...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板列表查询失败")
	}

	items := make([]entity.ProductTemplateListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.ProductTemplateListItem{
			ID:                row.ID,
			Title:             row.Title,
			Type:              row.TemplateType,
			TypeLabel:         productTemplateTypeLabel(row.TemplateType),
			IsShared:          row.IsShared,
			IsSharedLabel:     productTemplateSharedLabel(row.IsShared),
			AccountName:       row.AccountName,
			ValidateType:      row.ValidateType,
			ValidateTypeLabel: productTemplateValidateTypeTitles[row.ValidateType],
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
		})
	}
	return &adminapi.ProductTemplateListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

func (l *ProductTemplateLogic) getTemplate(ctx context.Context, id int64) (entity.ProductTemplate, error) {
	template := entity.ProductTemplate{}
	err := l.core.DB().GetCore().GetScan(ctx, &template, `
SELECT
    id,
    title,
    template_type,
    is_shared,
    account_name,
    validate_type,
    created_at,
    updated_at
FROM product_template
WHERE id = ?
`, id)
	return template, err
}

func (l *ProductTemplateLogic) loadTemplateIDs(ctx context.Context, ids []int64) ([]int64, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT id FROM product_template WHERE id IN (`+sqlPlaceholders(len(ids))+`)`, args...); err != nil {
		return nil, err
	}
	result := make([]int64, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.ID)
	}
	return result, nil
}
