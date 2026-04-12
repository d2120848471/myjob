package adminlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"strings"
	"time"

	authapi "myjob/api/auth"
	"myjob/internal/kernel"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"golang.org/x/crypto/bcrypt"
)

type AuthLogic struct{ core *kernel.Core }

func (l *AuthLogic) Login(ctx context.Context, req authapi.LoginReq, ip string) (map[string]any, *modelruntime.APIError) {
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		return nil, apiErr(http.StatusBadRequest, 400, "请输入用户名和密码")
	}
	user, err := l.core.GetUserByUsername(ctx, req.Username)
	if err != nil || user.IsDeleted == 1 {
		return nil, apiErr(http.StatusUnauthorized, 401, "账号或密码错误")
	}
	if user.Status != 1 {
		return nil, apiErr(http.StatusForbidden, 403, "账号已被禁用，请联系管理员")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apiErr(http.StatusUnauthorized, 401, "账号或密码错误")
	}
	if reason := kernel.SMSVerifyReason(user, ip); reason != "" {
		if strings.TrimSpace(user.Phone) == "" {
			return nil, apiErr(http.StatusForbidden, 403, "请联系管理员配置手机号")
		}
		loginToken := uuid.NewString()
		temp := modelruntime.TempLoginPayload{UserID: user.ID, IP: ip, Attempts: 0}
		if err = l.core.SaveTempLogin(ctx, loginToken, temp); err != nil {
			return nil, apiErr(http.StatusInternalServerError, 500, "登录临时态创建失败")
		}
		return map[string]any{"need_sms_verify": true, "login_token": loginToken, "phone": kernel.MaskPhone(user.Phone), "reason": reason}, nil
	}
	token, perms, err := l.core.IssueSession(ctx, user)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录失败")
	}
	if err = l.core.UpdateLoginState(ctx, user.ID, ip); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录状态更新失败")
	}
	_ = l.core.InsertLoginLog(ctx, user.ID, user.RealName, ip)
	return map[string]any{"need_sms_verify": false, "token": token, "user": l.core.BuildLoginUser(ctx, user), "permissions": perms}, nil
}

func (l *AuthLogic) LoginSMSSend(ctx context.Context, req authapi.LoginSMSSendReq) (map[string]any, *modelruntime.APIError) {
	if strings.TrimSpace(req.LoginToken) == "" {
		return nil, apiErr(http.StatusBadRequest, 400, "login_token不能为空")
	}
	temp, err := l.core.GetTempLogin(ctx, req.LoginToken)
	if err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "登录临时凭证已失效")
	}
	user, err := l.core.GetUserByID(ctx, temp.UserID)
	if err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "用户不存在")
	}
	lockKey := authlib.SMSSendLockKey(user.ID)
	if _, err = l.core.Redis().Get(ctx, lockKey).Result(); err == nil {
		return nil, apiErr(http.StatusTooManyRequests, 429, "请稍后再试")
	}
	cfg, err := l.core.LoadSMSConfig(ctx)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "短信配置读取失败")
	}
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
	payload := modelruntime.SMSCodePayload{LoginToken: req.LoginToken, Code: code}
	data, _ := json.Marshal(payload)
	if err = l.core.Redis().Set(ctx, authlib.SMSCodeKey(user.ID), data, time.Duration(cfg.ExpireMinutes)*time.Minute).Err(); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "验证码保存失败")
	}
	if err = l.core.Redis().Set(ctx, lockKey, "1", time.Duration(cfg.IntervalMinutes)*time.Minute).Err(); err != nil {
		_ = l.core.Redis().Del(ctx, authlib.SMSCodeKey(user.ID)).Err()
		return nil, apiErr(http.StatusInternalServerError, 500, "发送频控创建失败")
	}
	if err = l.core.Sender().SendLoginCode(ctx, user.Phone, code, cfg); err != nil {
		_ = l.core.Redis().Del(ctx, authlib.SMSCodeKey(user.ID), lockKey).Err()
		return nil, apiErr(http.StatusInternalServerError, 500, "短信发送失败")
	}
	return map[string]any{}, nil
}

func (l *AuthLogic) LoginSMSVerify(ctx context.Context, req authapi.LoginSMSVerifyReq) (map[string]any, *modelruntime.APIError) {
	if strings.TrimSpace(req.LoginToken) == "" || !kernel.SMSCodeRegexp().MatchString(req.SMSCode) {
		return nil, apiErr(http.StatusBadRequest, 400, "验证码格式错误")
	}
	temp, err := l.core.GetTempLogin(ctx, req.LoginToken)
	if err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "登录临时凭证已失效")
	}
	user, err := l.core.GetUserByID(ctx, temp.UserID)
	if err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "用户不存在")
	}
	rawCode, err := l.core.Redis().Get(ctx, authlib.SMSCodeKey(user.ID)).Result()
	if err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "验证码已失效")
	}
	var codePayload modelruntime.SMSCodePayload
	if err = json.Unmarshal([]byte(rawCode), &codePayload); err != nil {
		return nil, apiErr(http.StatusBadRequest, 400, "验证码已失效")
	}
	if codePayload.LoginToken != req.LoginToken || codePayload.Code != req.SMSCode {
		temp.Attempts++
		ttl, _ := l.core.Redis().TTL(ctx, authlib.TempLoginKey(req.LoginToken)).Result()
		if ttl <= 0 {
			ttl = time.Duration(l.core.Config().Auth.TempLoginTTLMin) * time.Minute
		}
		data, _ := json.Marshal(temp)
		_ = l.core.Redis().Set(ctx, authlib.TempLoginKey(req.LoginToken), data, ttl).Err()
		if temp.Attempts >= 5 {
			_ = l.core.Redis().Del(ctx, authlib.TempLoginKey(req.LoginToken), authlib.SMSCodeKey(user.ID)).Err()
			return nil, apiErr(http.StatusBadRequest, 400, "验证码错误，剩余 0 次机会")
		}
		return nil, apiErr(http.StatusBadRequest, 400, fmt.Sprintf("验证码错误，剩余 %d 次机会", 5-temp.Attempts))
	}
	_ = l.core.Redis().Del(ctx, authlib.TempLoginKey(req.LoginToken), authlib.SMSCodeKey(user.ID)).Err()
	token, perms, err := l.core.IssueSession(ctx, user)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录失败")
	}
	if err = l.core.UpdateLoginState(ctx, user.ID, temp.IP); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "登录状态更新失败")
	}
	_ = l.core.InsertLoginLog(ctx, user.ID, user.RealName, temp.IP)
	return map[string]any{"token": token, "user": l.core.BuildLoginUser(ctx, user), "permissions": perms}, nil
}

func (l *AuthLogic) Me(ctx context.Context, _ modelruntime.Principal, user kernel.AdminUser) (map[string]any, *modelruntime.APIError) {
	perms, err := l.core.LoadPermissions(ctx, user.GroupID)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "权限读取失败")
	}
	return map[string]any{"user": l.core.BuildLoginUser(ctx, user), "permissions": perms}, nil
}

func (l *AuthLogic) Logout(ctx context.Context, principal modelruntime.Principal, user kernel.AdminUser) (map[string]any, *modelruntime.APIError) {
	_ = l.core.RemoveSession(ctx, principal.JTI, user.ID)
	return map[string]any{}, nil
}
