package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"myjob/internal/app"
	"myjob/internal/consts"
)

func normalizeCustomerCompanyName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 100 {
		return "", apiErr(consts.CodeBadRequest, "请输入店铺或公司名称")
	}
	return name, nil
}

func normalizeCustomerPhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if !app.PhoneRegexp().MatchString(phone) {
		return "", apiErr(consts.CodeBadRequest, "手机号格式错误")
	}
	return phone, nil
}

func validateCustomerLoginPassword(password, confirm string) error {
	if !app.PasswordRegexp().MatchString(strings.TrimSpace(password)) || password != confirm {
		return apiErr(consts.CodeBadRequest, "登录密码格式错误")
	}
	return nil
}

func validateCustomerPayPassword(password, confirm string) error {
	if !app.PayPasswordRegexp().MatchString(strings.TrimSpace(password)) || password != confirm {
		return apiErr(consts.CodeBadRequest, "支付密码必须为6位数字")
	}
	return nil
}

func validateCustomerStatus(status int) error {
	if status != consts.StatusDisabled && status != consts.StatusEnabled {
		return apiErr(consts.CodeBadRequest, "状态错误")
	}
	return nil
}

func validateCustomerWritable(customer app.CustomerUser) error {
	if customer.IsDeleted == 1 {
		// 回收站客户只允许走恢复入口，避免通过普通写接口释放手机号占用或改动历史状态。
		return apiErr(consts.CodeConflict, "客户在回收站，请先恢复")
	}
	return nil
}

func (l *CustomerLogic) ensureCustomerPhoneAvailable(ctx context.Context, phone string, currentID int64) error {
	customer, err := l.core.GetCustomerByPhone(ctx, phone)
	if err == nil {
		if customer.ID != 0 && customer.ID != currentID {
			return apiErr(consts.CodeConflict, "手机号已存在")
		}
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return apiErr(consts.CodeInternalError, "客户查询失败")
}

func isCustomerUniquePhoneError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique")
}
