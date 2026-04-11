package app

import (
	"admin/utility/ipx"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (a *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, 400, "请输入用户名和密码")
		return
	}
	user, err := a.getUserByUsername(r.Context(), req.Username)
	if err != nil || user.IsDeleted == 1 {
		writeError(w, http.StatusUnauthorized, 401, "账号或密码错误")
		return
	}
	if user.Status != statusEnabled {
		writeError(w, http.StatusForbidden, 403, "账号已被禁用，请联系管理员")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, 401, "账号或密码错误")
		return
	}

	ip := requestIP(r)
	if reason := smsVerifyReason(user, ip); reason != "" {
		if strings.TrimSpace(user.Phone) == "" {
			writeError(w, http.StatusForbidden, 403, "请联系管理员配置手机号")
			return
		}
		loginToken := uuid.NewString()
		temp := tempLoginPayload{UserID: user.ID, IP: ip, Attempts: 0}
		if saveErr := a.saveTempLogin(r.Context(), loginToken, temp); saveErr != nil {
			writeError(w, http.StatusInternalServerError, 500, "登录临时态创建失败")
			return
		}
		writeSuccess(w, map[string]interface{}{
			"need_sms_verify": true,
			"login_token":     loginToken,
			"phone":           ipx.MaskPhone(user.Phone),
			"reason":          reason,
		})
		return
	}

	token, perms, issueErr := a.issueSession(r.Context(), user)
	if issueErr != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录失败")
		return
	}
	if err = a.updateLoginState(r.Context(), user.ID, ip); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录状态更新失败")
		return
	}
	_ = a.insertLoginLog(r.Context(), user.ID, user.RealName, ip)
	writeSuccess(w, map[string]interface{}{
		"need_sms_verify": false,
		"token":           token,
		"user":            a.buildLoginUser(r.Context(), user),
		"permissions":     perms,
	})
}

func (a *Application) handleLoginSMSSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LoginToken string `json:"login_token"`
	}
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.LoginToken) == "" {
		writeError(w, http.StatusBadRequest, 400, "login_token不能为空")
		return
	}
	temp, err := a.getTempLogin(r.Context(), req.LoginToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "登录临时凭证已失效")
		return
	}
	user, err := a.getUserByID(r.Context(), temp.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户不存在")
		return
	}
	lockKey := smsSendLockKey(user.ID)
	if _, err = a.redis.Get(r.Context(), lockKey).Result(); err == nil {
		writeError(w, http.StatusTooManyRequests, 429, "请稍后再试")
		return
	}
	cfg, err := a.loadSMSConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置读取失败")
		return
	}
	code := fmt.Sprintf("%06d", rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000000))
	payload := smsCodePayload{LoginToken: req.LoginToken, Code: code}
	data, _ := json.Marshal(payload)
	if err = a.redis.Set(r.Context(), smsCodeKey(user.ID), data, time.Duration(cfg.ExpireMinutes)*time.Minute).Err(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "验证码保存失败")
		return
	}
	if err = a.redis.Set(r.Context(), lockKey, "1", time.Duration(cfg.IntervalMinutes)*time.Minute).Err(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "发送频控创建失败")
		return
	}
	if err = a.sender.SendLoginCode(r.Context(), user.Phone, code, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信发送失败")
		return
	}
	writeSuccess(w, map[string]interface{}{})
}

func (a *Application) handleLoginSMSVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LoginToken string `json:"login_token"`
		SMSCode    string `json:"sms_code"`
	}
	if err := decodeJSON(r, &req); err != nil || strings.TrimSpace(req.LoginToken) == "" || !smsCodeRegexp.MatchString(req.SMSCode) {
		writeError(w, http.StatusBadRequest, 400, "验证码格式错误")
		return
	}
	temp, err := a.getTempLogin(r.Context(), req.LoginToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "登录临时凭证已失效")
		return
	}
	user, err := a.getUserByID(r.Context(), temp.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "用户不存在")
		return
	}
	rawCode, err := a.redis.Get(r.Context(), smsCodeKey(user.ID)).Result()
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, "验证码已失效")
		return
	}
	var codePayload smsCodePayload
	if err = json.Unmarshal([]byte(rawCode), &codePayload); err != nil {
		writeError(w, http.StatusBadRequest, 400, "验证码已失效")
		return
	}
	if codePayload.LoginToken != req.LoginToken || codePayload.Code != req.SMSCode {
		temp.Attempts++
		ttl, _ := a.redis.TTL(r.Context(), tempLoginKey(req.LoginToken)).Result()
		if ttl <= 0 {
			ttl = time.Duration(a.cfg.Auth.TempLoginTTLMin) * time.Minute
		}
		data, _ := json.Marshal(temp)
		_ = a.redis.Set(r.Context(), tempLoginKey(req.LoginToken), data, ttl).Err()
		if temp.Attempts >= 5 {
			_ = a.redis.Del(r.Context(), tempLoginKey(req.LoginToken), smsCodeKey(user.ID)).Err()
			writeError(w, http.StatusBadRequest, 400, "验证码错误，剩余 0 次机会")
			return
		}
		remaining := 5 - temp.Attempts
		writeError(w, http.StatusBadRequest, 400, fmt.Sprintf("验证码错误，剩余 %d 次机会", remaining))
		return
	}
	_ = a.redis.Del(r.Context(), tempLoginKey(req.LoginToken), smsCodeKey(user.ID)).Err()
	token, perms, issueErr := a.issueSession(r.Context(), user)
	if issueErr != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录失败")
		return
	}
	if err = a.updateLoginState(r.Context(), user.ID, temp.IP); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "登录状态更新失败")
		return
	}
	_ = a.insertLoginLog(r.Context(), user.ID, user.RealName, temp.IP)
	writeSuccess(w, map[string]interface{}{
		"token":       token,
		"user":        a.buildLoginUser(r.Context(), user),
		"permissions": perms,
	})
}

func (a *Application) handleMe(w http.ResponseWriter, r *http.Request, _ principal, user AdminUser) {
	perms, err := a.loadPermissions(r.Context(), user.GroupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "权限读取失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"user":        a.buildLoginUser(r.Context(), user),
		"permissions": perms,
	})
}

func (a *Application) handleLogout(w http.ResponseWriter, r *http.Request, p principal, user AdminUser) {
	_ = a.removeSession(r.Context(), p.JTI, user.ID)
	writeSuccess(w, map[string]interface{}{})
}
