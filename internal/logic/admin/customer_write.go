package adminlogic

import (
	"context"
	"fmt"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"golang.org/x/crypto/bcrypt"
)

// Add 新增客户账号，并写入操作日志。
func (l *CustomerLogic) Add(ctx context.Context, req *adminapi.CustomerCreateReq, actor app.AdminUser, ip string) (*adminapi.CustomerCreateRes, error) {
	companyName, err := normalizeCustomerCompanyName(req.CompanyName)
	if err != nil {
		return nil, err
	}
	phone, err := normalizeCustomerPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if err = validateCustomerStatus(req.Status); err != nil {
		return nil, err
	}
	if err = validateCustomerLoginPassword(req.Password, req.ConfirmPassword); err != nil {
		return nil, err
	}
	if err = validateCustomerPayPassword(req.PayPassword, req.ConfirmPayPassword); err != nil {
		return nil, err
	}
	if err = l.ensureCustomerPhoneAvailable(ctx, phone, 0); err != nil {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "密码加密失败")
	}
	payPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.PayPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "支付密码加密失败")
	}
	now := l.core.Now()
	result, err := l.core.DB().Exec(ctx, `
INSERT INTO customer_user (company_name, phone, password_hash, pay_password_hash, status, is_deleted, token_version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?)
	`, companyName, phone, string(passwordHash), string(payPasswordHash), req.Status, now, now)
	if err != nil {
		if isCustomerUniquePhoneError(err) {
			return nil, apiErr(consts.CodeConflict, "手机号已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "客户新增失败")
	}
	id, _ := result.LastInsertId()
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增客户：%s/%s", companyName, phone), ip)
	return &adminapi.CustomerCreateRes{ID: id}, nil
}

// Edit 编辑客户公司名、手机号和状态。
func (l *CustomerLogic) Edit(ctx context.Context, req *adminapi.CustomerUpdateReq, actor app.AdminUser, ip string) (*adminapi.CustomerUpdateRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if err = validateCustomerWritable(customer); err != nil {
		return nil, err
	}
	companyName, err := normalizeCustomerCompanyName(req.CompanyName)
	if err != nil {
		return nil, err
	}
	phone, err := normalizeCustomerPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if err = validateCustomerStatus(req.Status); err != nil {
		return nil, err
	}
	if err = l.ensureCustomerPhoneAvailable(ctx, phone, req.ID); err != nil {
		return nil, err
	}
	now := l.core.Now()
	if req.Status == consts.StatusDisabled && customer.Status != consts.StatusDisabled {
		_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET company_name = ?, phone = ?, status = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, companyName, phone, req.Status, now, req.ID)
	} else {
		_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET company_name = ?, phone = ?, status = ?, updated_at = ? WHERE id = ?`, companyName, phone, req.Status, now, req.ID)
	}
	if err != nil {
		if isCustomerUniquePhoneError(err) {
			return nil, apiErr(consts.CodeConflict, "手机号已存在")
		}
		return nil, apiErr(consts.CodeInternalError, "客户编辑失败")
	}
	if req.Status == consts.StatusDisabled && customer.Status != consts.StatusDisabled {
		_ = l.core.RemoveAllCustomerSessions(ctx, req.ID)
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑客户：%d", req.ID), ip)
	return &adminapi.CustomerUpdateRes{}, nil
}

// Status 启用或禁用客户；禁用时失效旧客户 token，避免被停用账号继续访问客户侧接口。
func (l *CustomerLogic) Status(ctx context.Context, req *adminapi.CustomerStatusReq, actor app.AdminUser, ip string) (*adminapi.CustomerStatusRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if err = validateCustomerWritable(customer); err != nil {
		return nil, err
	}
	if err = validateCustomerStatus(req.Status); err != nil {
		return nil, err
	}
	if req.Status == consts.StatusDisabled {
		_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET status = 0, token_version = token_version + 1, updated_at = ? WHERE id = ?`, l.core.Now(), req.ID)
		_ = l.core.RemoveAllCustomerSessions(ctx, req.ID)
	} else {
		_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET status = 1, updated_at = ? WHERE id = ?`, l.core.Now(), req.ID)
	}
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户状态更新失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("切换客户状态：客户ID=%d -> %d", req.ID, req.Status), ip)
	return &adminapi.CustomerStatusRes{}, nil
}

// Delete 软删除客户并失效旧客户 token；手机号仍保留唯一占用，避免历史账号归属被新账号覆盖。
func (l *CustomerLogic) Delete(ctx context.Context, req *adminapi.CustomerDeleteReq, actor app.AdminUser, ip string) (*adminapi.CustomerDeleteRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if err = validateCustomerWritable(customer); err != nil {
		return nil, err
	}
	now := l.core.Now()
	_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET status = 0, is_deleted = 1, deleted_at = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, now, now, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户删除失败")
	}
	_ = l.core.RemoveAllCustomerSessions(ctx, req.ID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除客户：客户ID=%d", req.ID), ip)
	return &adminapi.CustomerDeleteRes{}, nil
}

// Restore 从回收站恢复客户；恢复后保持禁用状态，避免误开放历史账号。
func (l *CustomerLogic) Restore(ctx context.Context, req *adminapi.CustomerRestoreReq, actor app.AdminUser, ip string) (*adminapi.CustomerRestoreRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if customer.IsDeleted != 1 {
		return nil, apiErr(consts.CodeConflict, "客户不在回收站")
	}
	_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET is_deleted = 0, status = 0, deleted_at = NULL, token_version = token_version + 1, updated_at = ? WHERE id = ?`, l.core.Now(), req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户恢复失败")
	}
	_ = l.core.RemoveAllCustomerSessions(ctx, req.ID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("恢复客户：客户ID=%d", req.ID), ip)
	return &adminapi.CustomerRestoreRes{}, nil
}

// ResetPassword 重置登录密码并失效旧客户 token；操作日志不记录明文密码。
func (l *CustomerLogic) ResetPassword(ctx context.Context, req *adminapi.CustomerPasswordResetReq, actor app.AdminUser, ip string) (*adminapi.CustomerPasswordResetRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if err = validateCustomerWritable(customer); err != nil {
		return nil, err
	}
	if err = validateCustomerLoginPassword(req.Password, req.ConfirmPassword); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "密码加密失败")
	}
	_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET password_hash = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, string(hash), l.core.Now(), req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录密码重置失败")
	}
	_ = l.core.RemoveAllCustomerSessions(ctx, req.ID)
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("重置客户登录密码：客户ID=%d", req.ID), ip)
	return &adminapi.CustomerPasswordResetRes{}, nil
}

// ResetPayPassword 重置支付密码；支付密码不参与登录态，所以不踢客户登录 token。
func (l *CustomerLogic) ResetPayPassword(ctx context.Context, req *adminapi.CustomerPayPasswordResetReq, actor app.AdminUser, ip string) (*adminapi.CustomerPayPasswordResetRes, error) {
	customer, err := l.core.GetCustomerByID(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "客户不存在")
	}
	if err = validateCustomerWritable(customer); err != nil {
		return nil, err
	}
	if err = validateCustomerPayPassword(req.PayPassword, req.ConfirmPayPassword); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.PayPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "支付密码加密失败")
	}
	_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET pay_password_hash = ?, updated_at = ? WHERE id = ?`, string(hash), l.core.Now(), req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "支付密码重置失败")
	}
	l.core.WriteOperation(ctx, actor, fmt.Sprintf("重置客户支付密码：客户ID=%d", req.ID), ip)
	return &adminapi.CustomerPayPasswordResetRes{}, nil
}
