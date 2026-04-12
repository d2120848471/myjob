package adminlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	v1 "myjob/api/admin/v1"
	"myjob/internal/app"
	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthLogic struct{ core *app.Core }

func (l *AuthLogic) Login(ctx context.Context, req *v1.AuthLoginReq, ip string) (*v1.AuthLoginRes, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return nil, apiErr(consts.CodeBadRequest, "请输入用户名和密码")
	}
	user, err := l.core.GetUserByUsername(ctx, req.Username)
	if err != nil || user.IsDeleted == 1 {
		return nil, apiErr(consts.CodeUnauthorized, "账号或密码错误")
	}
	if user.Status != 1 {
		return nil, apiErr(consts.CodeForbidden, "账号已被禁用，请联系管理员")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apiErr(consts.CodeUnauthorized, "账号或密码错误")
	}
	if reason := app.SMSVerifyReason(user, ip); reason != "" {
		if strings.TrimSpace(user.Phone) == "" {
			return nil, apiErr(consts.CodeForbidden, "请联系管理员配置手机号")
		}
		loginToken := uuid.NewString()
		temp := modelruntime.TempLoginPayload{UserID: user.ID, IP: ip, Attempts: 0}
		if err = l.core.SaveTempLogin(ctx, loginToken, temp); err != nil {
			return nil, apiErr(consts.CodeInternalError, "登录临时态创建失败")
		}
		return &v1.AuthLoginRes{
			NeedSMSVerify: true,
			LoginToken:    loginToken,
			Phone:         app.MaskPhone(user.Phone),
			Reason:        reason,
		}, nil
	}
	token, perms, err := l.core.IssueSession(ctx, user)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录失败")
	}
	if err = l.core.UpdateLoginState(ctx, user.ID, ip); err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录状态更新失败")
	}
	_ = l.core.InsertLoginLog(ctx, user.ID, user.RealName, ip)
	return &v1.AuthLoginRes{
		NeedSMSVerify: false,
		Token:         token,
		User:          ptrLoginUser(l.core.BuildLoginUser(ctx, user)),
		Permissions:   perms,
	}, nil
}

func (l *AuthLogic) LoginSMSSend(ctx context.Context, req *v1.AuthSMSSendReq) (*v1.AuthSMSSendRes, error) {
	if strings.TrimSpace(req.LoginToken) == "" {
		return nil, apiErr(consts.CodeBadRequest, "login_token不能为空")
	}
	temp, err := l.core.GetTempLogin(ctx, req.LoginToken)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "登录临时凭证已失效")
	}
	user, err := l.core.GetUserByID(ctx, temp.UserID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "用户不存在")
	}
	lockKey := authlib.SMSSendLockKey(user.ID)
	if exists, existsErr := l.core.Redis().GroupGeneric().Exists(ctx, lockKey); existsErr == nil && exists > 0 {
		return nil, apiErr(consts.CodeTooManyRequests, "请稍后再试")
	}
	cfg, err := l.core.LoadSMSConfig(ctx)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "短信配置读取失败")
	}
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
	payload := modelruntime.SMSCodePayload{LoginToken: req.LoginToken, Code: code}
	data, _ := json.Marshal(payload)
	if err = l.core.RedisSetString(ctx, authlib.SMSCodeKey(user.ID), string(data), time.Duration(cfg.ExpireMinutes)*time.Minute); err != nil {
		return nil, apiErr(consts.CodeInternalError, "验证码保存失败")
	}
	if err = l.core.RedisSetString(ctx, lockKey, "1", time.Duration(cfg.IntervalMinutes)*time.Minute); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.SMSCodeKey(user.ID))
		return nil, apiErr(consts.CodeInternalError, "发送频控创建失败")
	}
	if err = l.core.Sender().SendLoginCode(ctx, user.Phone, code, cfg); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.SMSCodeKey(user.ID), lockKey)
		return nil, apiErr(consts.CodeInternalError, "短信发送失败")
	}
	return &v1.AuthSMSSendRes{}, nil
}

func (l *AuthLogic) LoginSMSVerify(ctx context.Context, req *v1.AuthSMSVerifyReq) (*v1.AuthSMSVerifyRes, error) {
	if strings.TrimSpace(req.LoginToken) == "" || !app.SMSCodeRegexp().MatchString(req.SMSCode) {
		return nil, apiErr(consts.CodeBadRequest, "验证码格式错误")
	}
	temp, err := l.core.GetTempLogin(ctx, req.LoginToken)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "登录临时凭证已失效")
	}
	user, err := l.core.GetUserByID(ctx, temp.UserID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "用户不存在")
	}
	rawCode, err := l.core.RedisGetString(ctx, authlib.SMSCodeKey(user.ID))
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "验证码已失效")
	}
	var codePayload modelruntime.SMSCodePayload
	if err = json.Unmarshal([]byte(rawCode), &codePayload); err != nil {
		return nil, apiErr(consts.CodeBadRequest, "验证码已失效")
	}
	if codePayload.LoginToken != req.LoginToken || codePayload.Code != req.SMSCode {
		temp.Attempts++
		ttl, ttlErr := l.core.RedisTTL(ctx, authlib.TempLoginKey(req.LoginToken))
		if ttlErr != nil || ttl <= 0 {
			ttl = time.Duration(l.core.Config().Auth.TempLoginTTLMin) * time.Minute
		}
		data, _ := json.Marshal(temp)
		_ = l.core.RedisSetString(ctx, authlib.TempLoginKey(req.LoginToken), string(data), ttl)
		if temp.Attempts >= 5 {
			_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.TempLoginKey(req.LoginToken), authlib.SMSCodeKey(user.ID))
			return nil, apiErr(consts.CodeBadRequest, "验证码错误，剩余 0 次机会")
		}
		return nil, apiErr(consts.CodeBadRequest, fmt.Sprintf("验证码错误，剩余 %d 次机会", 5-temp.Attempts))
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.TempLoginKey(req.LoginToken), authlib.SMSCodeKey(user.ID))
	token, perms, err := l.core.IssueSession(ctx, user)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录失败")
	}
	if err = l.core.UpdateLoginState(ctx, user.ID, temp.IP); err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录状态更新失败")
	}
	_ = l.core.InsertLoginLog(ctx, user.ID, user.RealName, temp.IP)
	return &v1.AuthSMSVerifyRes{
		Token:       token,
		User:        l.core.BuildLoginUser(ctx, user),
		Permissions: perms,
	}, nil
}

func (l *AuthLogic) Me(ctx context.Context, _ modelruntime.Principal, user app.AdminUser) (*v1.AuthMeRes, error) {
	perms, err := l.core.LoadPermissions(ctx, user.GroupID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "权限读取失败")
	}
	return &v1.AuthMeRes{User: l.core.BuildLoginUser(ctx, user), Permissions: perms}, nil
}

func (l *AuthLogic) Logout(ctx context.Context, principal modelruntime.Principal, user app.AdminUser) (*v1.AuthSessionDeleteRes, error) {
	_ = l.core.RemoveSession(ctx, principal.JTI, user.ID)
	return &v1.AuthSessionDeleteRes{}, nil
}

func ptrLoginUser(user modelruntime.LoginUser) *modelruntime.LoginUser {
	return &user
}
