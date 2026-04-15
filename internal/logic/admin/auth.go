package adminlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthLogic 提供后台登录鉴权相关业务能力（账号密码登录 + 短信二次验证）。
type AuthLogic struct{ core *app.Core }

// Login 执行账号密码登录；当命中风控规则时返回 NeedSMSVerify=true 并下发 login_token。
func (l *AuthLogic) Login(ctx context.Context, req *adminapi.AuthLoginReq, ip string) (*adminapi.AuthLoginRes, error) {
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
		// 触发短信二次验证时，仅返回脱敏手机号，并在 Redis 写入临时登录态（带 TTL）。
		if strings.TrimSpace(user.Phone) == "" {
			return nil, apiErr(consts.CodeForbidden, "请联系管理员配置手机号")
		}
		loginToken := uuid.NewString()
		temp := modelruntime.TempLoginPayload{UserID: user.ID, IP: ip, Attempts: 0}
		if err = l.core.SaveTempLogin(ctx, loginToken, temp); err != nil {
			return nil, apiErr(consts.CodeInternalError, "登录临时态创建失败")
		}
		return &adminapi.AuthLoginRes{
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
	return &adminapi.AuthLoginRes{
		NeedSMSVerify: false,
		Token:         token,
		User:          ptrLoginUser(l.core.BuildLoginUser(ctx, user)),
		Permissions:   perms,
	}, nil
}

// LoginSMSSend 发送短信验证码（针对 Login 返回的 login_token）。
//
// 使用 Redis 频控锁限制发送间隔；发送失败会回滚验证码与频控锁，避免脏状态影响后续登录。
func (l *AuthLogic) LoginSMSSend(ctx context.Context, req *adminapi.AuthSMSSendReq) (*adminapi.AuthSMSSendRes, error) {
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
	// 先写验证码，再写频控锁：若写锁失败需要删除验证码，避免残留。
	if err = l.core.RedisSetString(ctx, authlib.SMSCodeKey(user.ID), string(data), time.Duration(cfg.ExpireMinutes)*time.Minute); err != nil {
		return nil, apiErr(consts.CodeInternalError, "验证码保存失败")
	}
	if err = l.core.RedisSetString(ctx, lockKey, "1", time.Duration(cfg.IntervalMinutes)*time.Minute); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.SMSCodeKey(user.ID))
		return nil, apiErr(consts.CodeInternalError, "发送频控创建失败")
	}
	// 调用供应商发送失败时，需要同时删除验证码与频控锁，避免用户被错误限流。
	if err = l.core.Sender().SendLoginCode(ctx, user.Phone, code, cfg); err != nil {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.SMSCodeKey(user.ID), lockKey)
		return nil, apiErr(consts.CodeInternalError, "短信发送失败")
	}
	return &adminapi.AuthSMSSendRes{}, nil
}

// LoginSMSVerify 校验短信验证码，通过后签发登录会话并清理临时登录态/验证码缓存。
func (l *AuthLogic) LoginSMSVerify(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (*adminapi.AuthSMSVerifyRes, error) {
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
		// 错误次数计数写回临时登录态，并尽量保留原 TTL，避免绕过次数限制。
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
	// 验证成功后清理临时态与验证码缓存，避免重复使用。
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.TempLoginKey(req.LoginToken), authlib.SMSCodeKey(user.ID))
	token, perms, err := l.core.IssueSession(ctx, user)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录失败")
	}
	if err = l.core.UpdateLoginState(ctx, user.ID, temp.IP); err != nil {
		return nil, apiErr(consts.CodeInternalError, "登录状态更新失败")
	}
	_ = l.core.InsertLoginLog(ctx, user.ID, user.RealName, temp.IP)
	return &adminapi.AuthSMSVerifyRes{
		Token:       token,
		User:        l.core.BuildLoginUser(ctx, user),
		Permissions: perms,
	}, nil
}

// Me 返回当前登录用户信息与权限码列表。
func (l *AuthLogic) Me(ctx context.Context, _ modelruntime.Principal, user app.AdminUser) (*adminapi.AuthMeRes, error) {
	perms, err := l.core.LoadPermissions(ctx, user.GroupID)
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "权限读取失败")
	}
	return &adminapi.AuthMeRes{User: l.core.BuildLoginUser(ctx, user), Permissions: perms}, nil
}

// Logout 退出登录（删除当前会话）。
func (l *AuthLogic) Logout(ctx context.Context, principal modelruntime.Principal, user app.AdminUser) (*adminapi.AuthSessionDeleteRes, error) {
	_ = l.core.RemoveSession(ctx, principal.JTI, user.ID)
	return &adminapi.AuthSessionDeleteRes{}, nil
}

func ptrLoginUser(user modelruntime.LoginUser) *modelruntime.LoginUser {
	return &user
}
