package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

// Add 新增平台账号，并写入操作日志。
func (l *SupplierPlatformLogic) Add(ctx context.Context, req *adminapi.SupplierPlatformCreateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformCreateRes, error) {
	normalized, err := l.normalizeSupplierPlatformInput(ctx, req.Name, req.Domain, req.BackupDomain, req.TypeID, req.SubjectID, req.HasTax, req.Status, consts.StatusEnabled, req.TokenID, req.SecretKey, req.ThresholdAmount, req.Sort, req.CrowdName)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}

	result, err := l.core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (
    name, provider_code, provider_name, type_id, subject_id, has_tax, status, domain, backup_domain,
    token_id, secret_key, extra_config, threshold_amount, sort, crowd_name,
    last_balance_status, last_balance_message, last_balance_trace_id, is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?, 0, '', '', 0, ?, ?)
`, normalized.Name, normalized.ProviderCode, normalized.ProviderName, normalized.TypeID, normalized.SubjectID, normalized.HasTax, normalized.Status, normalized.Domain, normalized.BackupDomain, normalized.TokenID, normalized.SecretKey, normalized.ThresholdAmount, normalized.Sort, normalized.CrowdName, l.core.Now(), l.core.Now())
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

// Edit 编辑平台账号，并写入操作日志。
func (l *SupplierPlatformLogic) Edit(ctx context.Context, req *adminapi.SupplierPlatformUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SupplierPlatformUpdateRes, error) {
	account, err := l.getSupplierPlatform(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErr(consts.CodeBadRequest, "平台账号不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号详情查询失败")
	}
	normalized, err := l.normalizeSupplierPlatformInput(ctx, req.Name, req.Domain, req.BackupDomain, req.TypeID, req.SubjectID, req.HasTax, req.Status, account.Status, req.TokenID, req.SecretKey, req.ThresholdAmount, req.Sort, req.CrowdName)
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

	now := l.core.Now()
	updateSQL := `
UPDATE supplier_platform_account
SET name = ?, provider_code = ?, provider_name = ?, type_id = ?, subject_id = ?, has_tax = ?, status = ?, domain = ?, backup_domain = ?,
    token_id = ?, secret_key = ?, threshold_amount = ?, sort = ?, crowd_name = ?, updated_at = ?`
	args := []any{normalized.Name, normalized.ProviderCode, normalized.ProviderName, normalized.TypeID, normalized.SubjectID, normalized.HasTax, normalized.Status, normalized.Domain, normalized.BackupDomain, normalized.TokenID, normalized.SecretKey, normalized.ThresholdAmount, normalized.Sort, normalized.CrowdName, now}
	if resetStatus {
		updateSQL += `, last_balance_status = 0, last_balance_message = '', last_balance_at = NULL, last_balance_trace_id = ''`
	}
	updateSQL += ` WHERE id = ? AND is_deleted = 0`
	args = append(args, req.ID)

	err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, execErr := tx.Exec(updateSQL, args...); execErr != nil {
			return execErr
		}
		if account.Status == consts.StatusEnabled && normalized.Status == consts.StatusDisabled {
			// 关闭平台后同步关停历史绑定，避免商品列表和后续履约继续把该平台当成可用渠道。
			if _, execErr := tx.Exec(`
UPDATE product_goods_channel_binding
SET dock_status = ?, updated_at = ?
WHERE platform_account_id = ? AND is_deleted = 0 AND dock_status <> ?
`, consts.StatusDisabled, now, req.ID, consts.StatusDisabled); execErr != nil {
				return execErr
			}
		}
		return nil
	})
	if err != nil {
		if isDuplicateDBError(err) {
			return nil, apiErr(consts.CodeConflict, "平台账号已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "平台账号编辑失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑第三方对接平台：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.SupplierPlatformUpdateRes{}, nil
}

// Delete 软删除平台账号（is_deleted=1），并写入操作日志。
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

func isDuplicateDBError(err error) bool {
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique")
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
