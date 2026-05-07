package customerlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func apiErr(code gcode.Code, message string) error {
	return gerror.NewCode(code, message)
}

func normalizePhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if !app.PhoneRegexp().MatchString(phone) {
		return "", apiErr(consts.CodeBadRequest, "手机号格式错误")
	}
	return phone, nil
}

func normalizeCompanyName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 100 {
		return "", apiErr(consts.CodeBadRequest, "请输入店铺或公司名称")
	}
	return name, nil
}

func validateLoginPassword(password, confirm string) error {
	if !app.PasswordRegexp().MatchString(strings.TrimSpace(password)) || password != confirm {
		return apiErr(consts.CodeBadRequest, "登录密码格式错误")
	}
	return nil
}

func validatePayPassword(password, confirm string) error {
	if !app.PayPasswordRegexp().MatchString(strings.TrimSpace(password)) || password != confirm {
		return apiErr(consts.CodeBadRequest, "支付密码必须为6位数字")
	}
	return nil
}

func normalizeSMSScene(scene string) (string, error) {
	scene = strings.TrimSpace(scene)
	switch scene {
	case app.CustomerSMSSceneRegister, app.CustomerSMSSceneForgotPassword:
		return scene, nil
	default:
		return "", apiErr(consts.CodeBadRequest, "验证码场景错误")
	}
}

func (l *AuthLogic) lookupCustomerByPhone(ctx context.Context, phone string) (app.CustomerUser, bool, error) {
	customer, err := l.core.GetCustomerByPhone(ctx, phone)
	if err == nil {
		return customer, customer.ID != 0, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return app.CustomerUser{}, false, nil
	}
	return app.CustomerUser{}, false, apiErr(consts.CodeInternalError, "客户查询失败")
}

func (l *AuthLogic) ensureCustomerPhoneAvailable(ctx context.Context, phone string) error {
	if _, found, err := l.lookupCustomerByPhone(ctx, phone); err != nil {
		return err
	} else if found {
		return apiErr(consts.CodeConflict, "手机号已注册")
	}
	return nil
}

func isCustomerUniquePhoneError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique")
}
