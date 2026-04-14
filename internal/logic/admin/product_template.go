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

const productTemplateTypeLocal = "local"

var productTemplateValidateTypes = []entity.ProductTemplateValidateTypeItem{
	{ID: 1, Title: "手机号"},
	{ID: 2, Title: "QQ号"},
	{ID: 3, Title: "手机号或者QQ号"},
	{ID: 4, Title: "邮箱"},
	{ID: 5, Title: "网址"},
	{ID: 6, Title: "纯数字"},
	{ID: 7, Title: "微信号"},
	{ID: 8, Title: "手机号或者微信号"},
	{ID: 9, Title: "QQ号或者微信号"},
	{ID: 10, Title: "手机号或者QQ号或微信号"},
	{ID: 11, Title: "禁止填写手机号"},
	{ID: 12, Title: "禁止填写邮箱"},
}

var productTemplateValidateTypeTitles = func() map[int]string {
	items := make(map[int]string, len(productTemplateValidateTypes))
	for _, item := range productTemplateValidateTypes {
		items[item.ID] = item.Title
	}
	return items
}()

type ProductTemplateLogic struct{ core *app.Core }

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

func (l *ProductTemplateLogic) Add(ctx context.Context, req *adminapi.ProductTemplateCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateCreateRes, error) {
	normalized, err := normalizeProductTemplateInput(req.Title, req.Type, req.IsShared, req.AccountName, req.ValidateType)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO product_template (
    title, template_type, is_shared, account_name, validate_type, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)
`, normalized.Title, normalized.Type, normalized.IsShared, normalized.AccountName, normalized.ValidateType, l.core.Now(), l.core.Now())
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品模板：%s", normalized.Title), ip)
	return &adminapi.ProductTemplateCreateRes{ID: id}, nil
}

func (l *ProductTemplateLogic) Edit(ctx context.Context, req *adminapi.ProductTemplateUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateUpdateRes, error) {
	if _, err := l.getTemplate(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	normalized, err := normalizeProductTemplateInput(req.Title, req.Type, req.IsShared, req.AccountName, req.ValidateType)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if _, err = l.core.DB().Exec(ctx, `
UPDATE product_template
SET title = ?, template_type = ?, is_shared = ?, account_name = ?, validate_type = ?, updated_at = ?
WHERE id = ?
`, normalized.Title, normalized.Type, normalized.IsShared, normalized.AccountName, normalized.ValidateType, l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品模板：%d -> %s", req.ID, normalized.Title), ip)
	return &adminapi.ProductTemplateUpdateRes{}, nil
}

func (l *ProductTemplateLogic) Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateDeleteRes, error) {
	template, err := l.getTemplate(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	goodsRefCount, err := countActiveGoodsReference(ctx, l.core.DB(), "product_template_id", req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除校验失败")
	}
	if goodsRefCount > 0 {
		return nil, apiErr(consts.CodeConflict, "该商品模板已被商品引用，请先处理关联商品")
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_template WHERE id = ?`, req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品模板：%d -> %s", req.ID, template.Title), ip)
	return &adminapi.ProductTemplateDeleteRes{}, nil
}

func (l *ProductTemplateLogic) BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateBatchDeleteRes, error) {
	if len(req.IDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "请至少选择一个商品模板")
	}
	ids, err := uniquePositiveInt64s(req.IDs, "模板ID")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	existing, err := l.loadTemplateIDs(ctx, ids)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板查询失败")
	}
	if len(existing) != len(ids) {
		return nil, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	referenced, err := hasActiveGoodsReferences(ctx, l.core.DB(), "product_template_id", ids)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板删除校验失败")
	}
	if referenced {
		return nil, apiErr(consts.CodeConflict, "选中的商品模板中存在被商品引用的记录，请先处理关联商品")
	}

	// 批量删除前先做全量存在性校验，避免部分删除成功、部分模板不存在导致结果不确定。
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if _, err = l.core.DB().Exec(ctx, `DELETE FROM product_template WHERE id IN (`+sqlPlaceholders(len(ids))+`)`, args...); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品模板批量删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量删除商品模板：%d项", len(ids)), ip)
	return &adminapi.ProductTemplateBatchDeleteRes{}, nil
}

func (l *ProductTemplateLogic) ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (*adminapi.ProductTemplateValidateTypeListRes, error) {
	items := make([]entity.ProductTemplateValidateTypeItem, 0, len(productTemplateValidateTypes))
	items = append(items, productTemplateValidateTypes...)
	return &adminapi.ProductTemplateValidateTypeListRes{List: items}, nil
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

type normalizedProductTemplateInput struct {
	Title        string
	Type         string
	IsShared     int
	AccountName  string
	ValidateType int
}

func normalizeProductTemplateInput(title, templateType string, isShared int, accountName string, validateType int) (normalizedProductTemplateInput, error) {
	title = strings.TrimSpace(title)
	accountName = strings.TrimSpace(accountName)
	if title == "" {
		return normalizedProductTemplateInput{}, fmt.Errorf("模板名称不能为空")
	}
	if accountName == "" {
		return normalizedProductTemplateInput{}, fmt.Errorf("充值账号名称不能为空")
	}
	if isShared != 0 && isShared != 1 {
		return normalizedProductTemplateInput{}, fmt.Errorf("共享状态错误")
	}
	normalizedType, err := normalizeProductTemplateType(templateType, true)
	if err != nil {
		return normalizedProductTemplateInput{}, err
	}
	if _, ok := productTemplateValidateTypeTitles[validateType]; !ok {
		return normalizedProductTemplateInput{}, fmt.Errorf("验证方式错误")
	}
	return normalizedProductTemplateInput{
		Title:        title,
		Type:         normalizedType,
		IsShared:     isShared,
		AccountName:  accountName,
		ValidateType: validateType,
	}, nil
}

func normalizeProductTemplateType(value string, defaultLocal bool) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		if defaultLocal {
			return productTemplateTypeLocal, nil
		}
		return "", nil
	}
	if value != productTemplateTypeLocal {
		return "", fmt.Errorf("模板类型错误")
	}
	return value, nil
}

func normalizeProductTemplateSharedFilter(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("共享状态错误")
	}
}

func productTemplateTypeLabel(value string) string {
	if value == productTemplateTypeLocal {
		return "本地模板"
	}
	return value
}

func productTemplateSharedLabel(value int) string {
	if value == 1 {
		return "共享"
	}
	return "不共享"
}
