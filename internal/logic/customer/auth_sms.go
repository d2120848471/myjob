package customerlogic

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	customerapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	modelruntime "myjob/internal/model/runtime"
)

const customerSMSMaxAttempts = 5

// SendSMS 发送客户注册或找回密码验证码，验证码按 scene+phone 隔离。
func (l *AuthLogic) SendSMS(ctx context.Context, req *customerapi.CustomerAuthSMSSendReq) (*customerapi.CustomerAuthSMSSendRes, error) {
	phone, err := normalizePhone(req.Phone)
	if err != nil {
		return nil, err
	}
	scene, err := normalizeSMSScene(req.Scene)
	if err != nil {
		return nil, err
	}
	customer, found, err := l.lookupCustomerByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if scene == app.CustomerSMSSceneRegister && found {
		return nil, apiErr(consts.CodeConflict, "手机号已注册")
	}
	if scene == app.CustomerSMSSceneForgotPassword {
		if !found || customer.IsDeleted == 1 || customer.Status != consts.StatusEnabled {
			return nil, apiErr(consts.CodeBadRequest, "该手机号不可找回密码")
		}
	}
	lockKey := app.CustomerSMSSendLockKey(scene, phone)
	attemptsKey := app.CustomerSMSAttemptsKey(scene, phone)
	if exists, existsErr := l.core.Redis().GroupGeneric().Exists(ctx, lockKey); existsErr == nil && exists > 0 {
		return nil, apiErr(consts.CodeTooManyRequests, "请稍后再试")
	}
	cfg, err := l.core.LoadSMSConfig(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "短信配置读取失败")
	}
	code, err := generateCustomerSMSCode()
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "验证码生成失败")
	}
	payload := modelruntime.CustomerSMSCodePayload{Scene: scene, Phone: phone, Code: code}
	data, _ := json.Marshal(payload)
	codeKey := app.CustomerSMSCodeKey(scene, phone)
	if err = l.core.RedisSetString(ctx, codeKey, string(data), time.Duration(cfg.ExpireMinutes)*time.Minute); err != nil {
		return nil, apiErr(consts.CodeInternalError, "验证码保存失败")
	}
	if err = l.core.RedisSetString(ctx, lockKey, "1", time.Duration(cfg.IntervalMinutes)*time.Minute); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, codeKey, attemptsKey)
		return nil, apiErr(consts.CodeInternalError, "发送频控创建失败")
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, attemptsKey)
	if err = l.core.Sender().SendCode(ctx, phone, code, cfg); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, codeKey, lockKey, attemptsKey)
		return nil, apiErr(consts.CodeInternalError, "短信发送失败")
	}
	return &customerapi.CustomerAuthSMSSendRes{}, nil
}

func (l *AuthLogic) consumeSMSCode(ctx context.Context, scene, phone, code string) error {
	if !app.SMSCodeRegexp().MatchString(code) {
		return apiErr(consts.CodeBadRequest, "验证码格式错误")
	}
	codeKey := app.CustomerSMSCodeKey(scene, phone)
	raw, err := l.core.RedisGetString(ctx, codeKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErr(consts.CodeBadRequest, "验证码已失效")
		}
		return apiErr(consts.CodeBadRequest, "验证码已失效")
	}
	var payload modelruntime.CustomerSMSCodePayload
	if err = json.Unmarshal([]byte(raw), &payload); err != nil {
		return apiErr(consts.CodeBadRequest, "验证码已失效")
	}
	if payload.Scene != scene || payload.Phone != phone || payload.Code != code {
		return l.recordCustomerSMSFailure(ctx, scene, phone, codeKey)
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, codeKey, app.CustomerSMSAttemptsKey(scene, phone))
	return nil
}

func (l *AuthLogic) recordCustomerSMSFailure(ctx context.Context, scene, phone, codeKey string) error {
	ttl, err := l.core.RedisTTL(ctx, codeKey)
	if err != nil || ttl <= 0 {
		return apiErr(consts.CodeBadRequest, "验证码已失效")
	}
	attemptsKey := app.CustomerSMSAttemptsKey(scene, phone)
	attempts, err := l.core.Redis().GroupString().Incr(ctx, attemptsKey)
	if err != nil {
		return apiErr(consts.CodeInternalError, "验证码校验失败")
	}
	if _, err = l.core.Redis().GroupGeneric().Expire(ctx, attemptsKey, int64(ttl.Seconds())); err != nil {
		return apiErr(consts.CodeInternalError, "验证码校验失败")
	}
	if attempts >= customerSMSMaxAttempts {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, codeKey, attemptsKey)
		return apiErr(consts.CodeBadRequest, "验证码错误，剩余 0 次机会")
	}
	return apiErr(consts.CodeBadRequest, fmt.Sprintf("验证码错误，剩余 %d 次机会", customerSMSMaxAttempts-attempts))
}

func generateCustomerSMSCode() (string, error) {
	number, err := crand.Int(crand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}
