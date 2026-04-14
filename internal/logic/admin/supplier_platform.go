package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

type PlatformDeleteHook interface {
	BeforeDelete(ctx context.Context, platformID int64) error
}

type noopPlatformDeleteHook struct{}

func (noopPlatformDeleteHook) BeforeDelete(context.Context, int64) error { return nil }

type SupplierPlatformLogic struct {
	core       *app.Core
	httpClient *http.Client
	deleteHook PlatformDeleteHook
}

func NewSupplierPlatformLogic(core *app.Core) *SupplierPlatformLogic {
	return &SupplierPlatformLogic{
		core:       core,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		deleteHook: noopPlatformDeleteHook{},
	}
}

func (l *SupplierPlatformLogic) TypeList(ctx context.Context, _ *adminapi.SupplierPlatformTypeListReq) (*adminapi.SupplierPlatformTypeListRes, error) {
	items := make([]entity.SupplierPlatformTypeItem, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &items, `SELECT id, type_name, default_provider_code AS provider_code, status, sort, created_at, updated_at FROM supplier_platform_type WHERE status = 1 ORDER BY sort ASC, id ASC`); err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台类型查询失败")
	}
	return &adminapi.SupplierPlatformTypeListRes{List: items}, nil
}

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

func (l *SupplierPlatformLogic) Add(ctx context.Context, req *adminapi.SupplierPlatformCreateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformCreateRes, error) {
	normalized, err := l.normalizeSupplierPlatformInput(ctx, req.Name, req.Domain, req.BackupDomain, req.TypeID, req.SubjectID, req.HasTax, req.TokenID, req.SecretKey, req.ThresholdAmount, req.Sort, req.CrowdName)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	result, err := l.core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (
    name, provider_code, provider_name, type_id, subject_id, has_tax, domain, backup_domain,
    token_id, secret_key, extra_config, threshold_amount, sort, crowd_name,
    last_balance_status, last_balance_message, last_balance_trace_id, is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?, 0, '', '', 0, ?, ?)
`, normalized.Name, normalized.ProviderCode, normalized.ProviderName, normalized.TypeID, normalized.SubjectID, normalized.HasTax, normalized.Domain, normalized.BackupDomain, normalized.TokenID, normalized.SecretKey, normalized.ThresholdAmount, normalized.Sort, normalized.CrowdName, l.core.Now(), l.core.Now())
	if err != nil {
		if isDuplicateDBError(err) {
			return nil, apiErr(consts.CodeConflict, "平台账号已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增第三方对接平台：%s", normalized.Name), ip)
	return &adminapi.SupplierPlatformCreateRes{ID: id}, nil
}

func (l *SupplierPlatformLogic) Edit(ctx context.Context, req *adminapi.SupplierPlatformUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformUpdateRes, error) {
	account, err := l.getSupplierPlatform(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "平台账号不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号详情查询失败")
	}
	normalized, err := l.normalizeSupplierPlatformInput(ctx, req.Name, req.Domain, req.BackupDomain, req.TypeID, req.SubjectID, req.HasTax, req.TokenID, req.SecretKey, req.ThresholdAmount, req.Sort, req.CrowdName)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	resetStatus := account.Domain != normalized.Domain ||
		account.BackupDomain != normalized.BackupDomain ||
		account.TypeID != normalized.TypeID ||
		account.SubjectID != normalized.SubjectID ||
		account.HasTax != normalized.HasTax ||
		account.TokenID != normalized.TokenID ||
		account.SecretKey != normalized.SecretKey

	updateSQL := `
UPDATE supplier_platform_account
SET name = ?, provider_code = ?, provider_name = ?, type_id = ?, subject_id = ?, has_tax = ?, domain = ?, backup_domain = ?,
    token_id = ?, secret_key = ?, threshold_amount = ?, sort = ?, crowd_name = ?, updated_at = ?`
	args := []any{normalized.Name, normalized.ProviderCode, normalized.ProviderName, normalized.TypeID, normalized.SubjectID, normalized.HasTax, normalized.Domain, normalized.BackupDomain, normalized.TokenID, normalized.SecretKey, normalized.ThresholdAmount, normalized.Sort, normalized.CrowdName, l.core.Now()}
	if resetStatus {
		updateSQL += `, last_balance_status = 0, last_balance_message = '', last_balance_at = NULL, last_balance_trace_id = ''`
	}
	updateSQL += ` WHERE id = ? AND is_deleted = 0`
	args = append(args, req.ID)

	if _, err = l.core.DB().Exec(ctx, updateSQL, args...); err != nil {
		if isDuplicateDBError(err) {
			return nil, apiErr(consts.CodeConflict, "平台账号已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑第三方对接平台：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.SupplierPlatformUpdateRes{}, nil
}

func (l *SupplierPlatformLogic) Delete(ctx context.Context, req *adminapi.SupplierPlatformDeleteReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformDeleteRes, error) {
	account, err := l.getSupplierPlatform(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "平台账号不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号详情查询失败")
	}
	if err = l.deleteHook.BeforeDelete(ctx, req.ID); err != nil {
		return nil, apiErr(consts.CodeConflict, err.Error())
	}
	// 软删时同步改写 token_id，释放唯一键占位，避免同主体同平台账号被重新创建后再次删除时撞唯一索引。
	archivedTokenID := archivedSupplierTokenID(account.TokenID, req.ID)
	if _, err = l.core.DB().Exec(ctx, `UPDATE supplier_platform_account SET token_id = ?, is_deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`, archivedTokenID, l.core.Now(), l.core.Now(), req.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "平台账号删除失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除第三方对接平台：%d -> %s", req.ID, account.Name), ip)
	return &adminapi.SupplierPlatformDeleteRes{}, nil
}

type normalizedSupplierPlatformInput struct {
	Name            string
	ProviderCode    string
	ProviderName    string
	TypeID          int
	SubjectID       int64
	HasTax          int
	Domain          string
	BackupDomain    string
	TokenID         string
	SecretKey       string
	ThresholdAmount string
	Sort            int
	CrowdName       string
}

func (l *SupplierPlatformLogic) normalizeSupplierPlatformInput(ctx context.Context, name, domain, backupDomain string, typeID int, subjectID int64, hasTax int, tokenID, secretKey, thresholdAmount string, sortValue int, crowdName string) (normalizedSupplierPlatformInput, error) {
	name = strings.TrimSpace(name)
	domain = strings.TrimSpace(domain)
	backupDomain = strings.TrimSpace(backupDomain)
	tokenID = strings.TrimSpace(tokenID)
	secretKey = strings.TrimSpace(secretKey)
	crowdName = strings.TrimSpace(crowdName)
	if name == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台名称不能为空")
	}
	if domain == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("主域名不能为空")
	}
	if backupDomain == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("备用域名不能为空")
	}
	if typeID <= 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台类型不能为空")
	}
	if subjectID <= 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("主体不能为空")
	}
	if hasTax != 0 && hasTax != 1 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("含税值错误")
	}
	if tokenID == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台账号ID不能为空")
	}
	if secretKey == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台密钥不能为空")
	}
	if len(name) > 128 || len(tokenID) > 128 || len(secretKey) > 255 || len(crowdName) > 128 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台参数长度超出限制")
	}
	if sortValue < 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("排序值不能小于0")
	}
	if err := validateSupplierDomain(domain); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	if err := validateSupplierDomain(backupDomain); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	domain = normalizeSupplierDomain(domain)
	backupDomain = normalizeSupplierDomain(backupDomain)
	if err := l.ensureSupplierPlatformTypeExists(ctx, typeID); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	if err := l.ensureSubjectExists(ctx, subjectID); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(thresholdAmount))
	if err != nil || amount.IsNegative() {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("余额阈值错误")
	}
	providerCode, providerName, err := supplierprovider.ResolveProvider(typeID, domain, backupDomain, name)
	if err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	return normalizedSupplierPlatformInput{
		Name:            name,
		ProviderCode:    providerCode,
		ProviderName:    providerName,
		TypeID:          typeID,
		SubjectID:       subjectID,
		HasTax:          hasTax,
		Domain:          domain,
		BackupDomain:    backupDomain,
		TokenID:         tokenID,
		SecretKey:       secretKey,
		ThresholdAmount: amount.StringFixed(4),
		Sort:            sortValue,
		CrowdName:       crowdName,
	}, nil
}

func (l *SupplierPlatformLogic) getSupplierPlatform(ctx context.Context, id int64) (entity.SupplierPlatformAccount, error) {
	rows, err := l.core.DB().GetCore().GetAll(ctx, `
SELECT
    id, name, provider_code, provider_name, type_id, subject_id, has_tax, domain, backup_domain,
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

func (l *SupplierPlatformLogic) ensureSupplierPlatformTypeExists(ctx context.Context, typeID int) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM supplier_platform_type WHERE id = ? AND status = 1`, typeID)
	if err != nil {
		return fmt.Errorf("平台类型查询失败")
	}
	if count.Int() == 0 {
		return fmt.Errorf("平台类型不存在")
	}
	return nil
}

func (l *SupplierPlatformLogic) ensureSubjectExists(ctx context.Context, subjectID int64) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE id = ?`, subjectID)
	if err != nil {
		return fmt.Errorf("主体查询失败")
	}
	if count.Int() == 0 {
		return fmt.Errorf("主体不存在")
	}
	return nil
}

func normalizeSupplierDomain(value string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(value, "https://"), "http://"), "/"))
}

func validateSupplierDomain(value string) error {
	lower := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return fmt.Errorf("域名不能带协议头")
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("域名不能包含路径")
	}
	return nil
}

func normalizeSupplierTaxFilter(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("含税筛选值错误")
	}
}

func normalizeSupplierConnectStatus(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	case "2":
		return 2, true, nil
	default:
		return 0, false, fmt.Errorf("对接状态筛选值错误")
	}
}

func parseExtraConfig(raw string) (map[string]any, error) {
	result := map[string]any{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func nullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func supplierPlatformAccountFromRecord(row gdb.Record) entity.SupplierPlatformAccount {
	return entity.SupplierPlatformAccount{
		ID:                 row["id"].Int64(),
		Name:               row["name"].String(),
		ProviderCode:       row["provider_code"].String(),
		ProviderName:       row["provider_name"].String(),
		TypeID:             row["type_id"].Int(),
		SubjectID:          row["subject_id"].Int64(),
		HasTax:             row["has_tax"].Int(),
		Domain:             row["domain"].String(),
		BackupDomain:       row["backup_domain"].String(),
		TokenID:            row["token_id"].String(),
		SecretKey:          row["secret_key"].String(),
		ExtraConfig:        row["extra_config"].String(),
		ThresholdAmount:    formatMoney(row["threshold_amount"].String()),
		Sort:               row["sort"].Int(),
		CrowdName:          row["crowd_name"].String(),
		LastBalance:        nullableStringFromRecord(row, "last_balance"),
		LastBalanceStatus:  row["last_balance_status"].Int(),
		LastBalanceMessage: row["last_balance_message"].String(),
		LastBalanceAt:      nullableTimeFromRecord(row, "last_balance_at"),
		LastBalanceTraceID: row["last_balance_trace_id"].String(),
		IsDeleted:          row["is_deleted"].Int(),
		DeletedAt:          nullableTimeFromRecord(row, "deleted_at"),
		CreatedAt:          parseRecordTime(row, "created_at"),
		UpdatedAt:          parseRecordTime(row, "updated_at"),
	}
}

func supplierPlatformListItemFromRecord(row gdb.Record) entity.SupplierPlatformListItem {
	lastBalance := formatMoney(nullableStringFromRecord(row, "last_balance").String)
	threshold := formatMoney(row["threshold_amount"].String())
	connectStatus := row["last_balance_status"].Int()
	return entity.SupplierPlatformListItem{
		ID:                 row["id"].Int64(),
		Name:               row["name"].String(),
		Domain:             row["domain"].String(),
		BackupDomain:       row["backup_domain"].String(),
		ProviderCode:       row["provider_code"].String(),
		ProviderName:       row["provider_name"].String(),
		TypeID:             row["type_id"].Int(),
		TypeName:           row["type_name"].String(),
		SubjectID:          row["subject_id"].Int64(),
		SubjectName:        row["subject_name"].String(),
		HasTax:             row["has_tax"].Int(),
		LastBalance:        lastBalance,
		ThresholdAmount:    threshold,
		BalanceWarning:     balanceWarning(lastBalance, threshold),
		ConnectStatus:      connectStatus,
		ConnectStatusText:  connectStatusText(connectStatus),
		LastBalanceMessage: row["last_balance_message"].String(),
		LastBalanceAt:      formatNullableTime(nullableTimeFromRecord(row, "last_balance_at")),
		Sort:               row["sort"].Int(),
		CrowdName:          row["crowd_name"].String(),
	}
}

func nullableStringFromRecord(row gdb.Record, key string) sql.NullString {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.TrimSpace(value.String()), Valid: true}
}

func nullableTimeFromRecord(row gdb.Record, key string) sql.NullTime {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullTime{}
	}
	parsed, ok := appTimeFromValue(value.Val())
	if !ok {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}

func parseRecordTime(row gdb.Record, key string) time.Time {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return time.Time{}
	}
	parsed, ok := appTimeFromValue(value.Val())
	if !ok {
		return time.Time{}
	}
	return parsed
}

func appTimeFromValue(raw any) (time.Time, bool) {
	switch value := raw.(type) {
	case time.Time:
		if value.IsZero() {
			return time.Time{}, false
		}
		return value, true
	case string:
		return parseRecordTimeString(value)
	case []byte:
		return parseRecordTimeString(string(value))
	case interface{ String() string }:
		return parseRecordTimeString(value.String())
	default:
		return time.Time{}, false
	}
}

func parseRecordTimeString(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func formatNullableTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02 15:04:05")
}

func connectStatusText(status int) string {
	switch status {
	case 1:
		return "正常"
	case 2:
		return "异常"
	default:
		return "未验证"
	}
}

func balanceWarning(lastBalance, threshold string) int {
	if strings.TrimSpace(lastBalance) == "" {
		return 0
	}
	thresholdAmount, err := decimal.NewFromString(strings.TrimSpace(threshold))
	if err != nil || thresholdAmount.LessThanOrEqual(decimal.Zero) {
		return 0
	}
	lastAmount, err := decimal.NewFromString(strings.TrimSpace(lastBalance))
	if err != nil {
		return 0
	}
	if lastAmount.LessThan(thresholdAmount) {
		return 1
	}
	return 0
}

func isDuplicateDBError(err error) bool {
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique")
}

func formatMoney(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return value
	}
	return amount.StringFixed(4)
}

func archivedSupplierTokenID(tokenID string, id int64) string {
	suffix := fmt.Sprintf("__deleted__%d", id)
	maxPrefixLength := 128 - len(suffix)
	if maxPrefixLength < 0 {
		maxPrefixLength = 0
	}
	tokenID = strings.TrimSpace(tokenID)
	if len(tokenID) > maxPrefixLength {
		tokenID = tokenID[:maxPrefixLength]
	}
	return tokenID + suffix
}
