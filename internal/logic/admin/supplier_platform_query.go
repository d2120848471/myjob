package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"
)

// TypeList 查询可用的平台类型字典列表。
func (l *SupplierPlatformLogic) TypeList(ctx context.Context, _ *adminapi.SupplierPlatformTypeListReq) (*adminapi.SupplierPlatformTypeListRes, error) {
	items := make([]entity.SupplierPlatformTypeItem, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, type_name, default_provider_code AS provider_code, status, sort, created_at, updated_at FROM supplier_platform_type WHERE status = 1 ORDER BY sort ASC, id ASC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台类型查询失败")
	}
	return &adminapi.SupplierPlatformTypeListRes{List: items}, nil
}

// List 分页查询平台账号列表，支持关键词/类型/主体/含税/连接状态筛选。
func (l *SupplierPlatformLogic) List(ctx context.Context, req *adminapi.SupplierPlatformListReq) (*adminapi.SupplierPlatformListRes, error) {
	page, pageSize := app.ParsePagination(req.Page, req.PageSize)
	hasTax, hasTaxFilter, err := normalizeSupplierTaxFilter(req.HasTax)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	connectStatus, hasConnectFilter, err := normalizeSupplierConnectStatus(req.ConnectStatus)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	conditions := []string{"a.is_deleted = 0"}
	args := make([]any, 0, 12)
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		conditions = append(conditions, "a.name LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	if req.TypeID > 0 {
		conditions = append(conditions, "a.type_id = ?")
		args = append(args, req.TypeID)
	}
	if req.SubjectID > 0 {
		conditions = append(conditions, "a.subject_id = ?")
		args = append(args, req.SubjectID)
	}
	if hasTaxFilter {
		conditions = append(conditions, "a.has_tax = ?")
		args = append(args, hasTax)
	}
	if hasConnectFilter {
		conditions = append(conditions, "a.last_balance_status = ?")
		args = append(args, connectStatus)
	}

	whereClause := strings.Join(conditions, " AND ")
	total, err := l.core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM supplier_platform_account a
JOIN supplier_platform_type t ON t.id = a.type_id
JOIN admin_subject s ON s.id = a.subject_id
WHERE `+whereClause, args...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台账号列表查询失败")
	}

	queryArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    a.id,
    a.name,
    a.domain,
    a.backup_domain,
    a.provider_code,
    a.provider_name,
    a.type_id,
    t.type_name,
    a.subject_id,
    s.name AS subject_name,
    a.has_tax,
    a.status,
    a.last_balance,
    a.threshold_amount,
    a.last_balance_status,
    a.last_balance_message,
    a.last_balance_at,
    a.sort,
    a.crowd_name
FROM supplier_platform_account a
JOIN supplier_platform_type t ON t.id = a.type_id
JOIN admin_subject s ON s.id = a.subject_id
WHERE `+whereClause+`
ORDER BY a.sort ASC, a.id DESC
LIMIT ? OFFSET ?
`, queryArgs...)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台账号列表查询失败")
	}

	items := make([]entity.SupplierPlatformListItem, 0, len(rows))
	for _, row := range rows {
		item := supplierPlatformListItemFromRecord(row)
		items = append(items, item)
	}
	return &adminapi.SupplierPlatformListRes{
		List:       items,
		Pagination: adminapi.PaginationRes{Page: page, PageSize: pageSize, Total: total.Int()},
	}, nil
}

// Detail 查询平台账号详情（敏感字段会脱敏返回）。
func (l *SupplierPlatformLogic) Detail(ctx context.Context, req *adminapi.SupplierPlatformDetailReq) (*adminapi.SupplierPlatformDetailRes, error) {
	account, err := l.getSupplierPlatform(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "平台账号不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号详情查询失败")
	}
	extraConfig, err := parseExtraConfig(account.ExtraConfig)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台扩展配置解析失败")
	}
	return &adminapi.SupplierPlatformDetailRes{
		ID:              account.ID,
		Name:            account.Name,
		Domain:          account.Domain,
		BackupDomain:    account.BackupDomain,
		TypeID:          account.TypeID,
		SubjectID:       account.SubjectID,
		HasTax:          account.HasTax,
		Status:          account.Status,
		TokenID:         account.TokenID,
		SecretKey:       account.SecretKey,
		ThresholdAmount: formatMoney(account.ThresholdAmount),
		Sort:            account.Sort,
		CrowdName:       account.CrowdName,
		ProviderCode:    account.ProviderCode,
		ProviderName:    account.ProviderName,
		ExtraConfig:     extraConfig,
	}, nil
}

func (l *SupplierPlatformLogic) getSupplierPlatform(ctx context.Context, id int64) (entity.SupplierPlatformAccount, error) {
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    id, name, provider_code, provider_name, type_id, subject_id, has_tax, status, domain, backup_domain,
    token_id, secret_key, extra_config, threshold_amount, sort, crowd_name, last_balance,
    last_balance_status, last_balance_message, last_balance_at, last_balance_trace_id, is_deleted,
    deleted_at, created_at, updated_at
FROM supplier_platform_account
WHERE id = ? AND is_deleted = 0
`, id)
	if err != nil {
		return entity.SupplierPlatformAccount{}, err
	}
	if len(rows) == 0 {
		return entity.SupplierPlatformAccount{}, sql.ErrNoRows
	}
	return supplierPlatformAccountFromRecord(rows[0]), nil
}
