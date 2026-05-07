package customerlogic

import (
	"context"
	"database/sql"

	customerapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"

	"golang.org/x/crypto/bcrypt"
)

// Register 注册客户账号，注册成功后直接签发客户 token。
func (l *AuthLogic) Register(ctx context.Context, req *customerapi.CustomerRegisterReq, ip string) (*customerapi.CustomerRegisterRes, error) {
	companyName, err := normalizeCompanyName(req.CompanyName)
	if err != nil {
		return nil, err
	}
	phone, err := normalizePhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if err = validateLoginPassword(req.Password, req.ConfirmPassword); err != nil {
		return nil, err
	}
	if err = validatePayPassword(req.PayPassword, req.ConfirmPayPassword); err != nil {
		return nil, err
	}
	if err = l.ensureCustomerPhoneAvailable(ctx, phone); err != nil {
		return nil, err
	}
	if err = l.consumeSMSCode(ctx, app.CustomerSMSSceneRegister, phone, req.SMSCode); err != nil {
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
INSERT INTO customer_user (
    company_name, phone, password_hash, pay_password_hash, status, is_deleted,
    last_login_ip, last_login_at, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, 1, 0, ?, ?, 0, ?, ?)
	`, companyName, phone, string(passwordHash), string(payPasswordHash), ip, now, now, now)
	if err != nil {
		if isCustomerUniquePhoneError(err) {
			return nil, apiErr(consts.CodeConflict, "手机号已注册")
		}
		return nil, apiErr(consts.CodeInternalError, "客户注册失败")
	}
	id, _ := result.LastInsertId()
	customer, err := l.core.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "客户读取失败")
	}
	token, err := l.core.IssueCustomerSession(ctx, customer)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录失败")
	}
	return &customerapi.CustomerRegisterRes{Token: token, Customer: l.core.BuildCustomerLoginUser(customer)}, nil
}

// Login 使用手机号和登录密码登录客户账号。
func (l *AuthLogic) Login(ctx context.Context, req *customerapi.CustomerLoginReq, ip string) (*customerapi.CustomerLoginRes, error) {
	phone, err := normalizePhone(req.Phone)
	if err != nil {
		return nil, err
	}
	customer, found, err := l.lookupCustomerByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if !found || customer.IsDeleted == 1 {
		return nil, apiErr(consts.CodeUnauthorized, "账号或密码错误")
	}
	if customer.Status != consts.StatusEnabled {
		return nil, apiErr(consts.CodeForbidden, "账号已被禁用，请联系客服")
	}
	if bcrypt.CompareHashAndPassword([]byte(customer.PasswordHash), []byte(req.Password)) != nil {
		return nil, apiErr(consts.CodeUnauthorized, "账号或密码错误")
	}
	now := l.core.Now()
	_, _ = l.core.DB().Exec(ctx, `UPDATE customer_user SET last_login_ip = ?, last_login_at = ?, updated_at = ? WHERE id = ?`, ip, now, now, customer.ID)
	customer.LastLoginIP = ip
	customer.LastLoginAt = sql.NullTime{Time: now, Valid: true}
	token, err := l.core.IssueCustomerSession(ctx, customer)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录失败")
	}
	return &customerapi.CustomerLoginRes{Token: token, Customer: l.core.BuildCustomerLoginUser(customer)}, nil
}

// ForgotPassword 通过短信验证码重置客户登录密码，并失效旧 token。
func (l *AuthLogic) ForgotPassword(ctx context.Context, req *customerapi.CustomerForgotPasswordReq) (*customerapi.CustomerForgotPasswordRes, error) {
	phone, err := normalizePhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if err = validateLoginPassword(req.Password, req.ConfirmPassword); err != nil {
		return nil, err
	}
	customer, found, err := l.lookupCustomerByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if !found || customer.IsDeleted == 1 || customer.Status != consts.StatusEnabled {
		return nil, apiErr(consts.CodeBadRequest, "该手机号不可找回密码")
	}
	if err = l.consumeSMSCode(ctx, app.CustomerSMSSceneForgotPassword, phone, req.SMSCode); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "密码加密失败")
	}
	_, err = l.core.DB().Exec(ctx, `UPDATE customer_user SET password_hash = ?, token_version = token_version + 1, updated_at = ? WHERE id = ?`, string(hash), l.core.Now(), customer.ID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "密码重置失败")
	}
	_ = l.core.RemoveAllCustomerSessions(ctx, customer.ID)
	return &customerapi.CustomerForgotPasswordRes{}, nil
}
