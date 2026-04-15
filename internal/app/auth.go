package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/errors/gerror"
)

// AuthenticateRequest 从 Authorization Header 解析并校验登录态，返回当前 principal 与用户信息。
//
// 校验链路：
// - Bearer JWT（校验签名/过期，解析出 jti/user_id/token_version）
// - Redis Session（校验 jti 是否仍有效、token_version 是否一致）
// - 用户/用户组状态（禁用/删除/用户组禁用等）
func (c *Core) AuthenticateRequest(ctx context.Context, authz string) (modelruntime.Principal, AdminUser, error) {
	authz = strings.TrimSpace(authz)
	if !strings.HasPrefix(authz, "Bearer ") {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	// tokenString 只包含 JWT 本体，不包含 Bearer 前缀。
	tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	claims, err := authlib.ParseToken(c.cfg.Auth.JWTSecret, tokenString)
	if err != nil || claims.ID == "" {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	// SessionPayload 存在 Redis，用于实现服务端可控的失效（踢下线、改密码等）。
	var session modelruntime.SessionPayload
	raw, redisErr := c.RedisGetString(ctx, authlib.SessionKey(claims.ID))
	if redisErr != nil {
		if redisErr == sql.ErrNoRows {
			return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
		}
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	if err = json.Unmarshal([]byte(raw), &session); err != nil {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	// 读取用户最新状态（禁用/删除等），避免仅依赖 token。
	user, queryErr := c.GetUserByID(ctx, claims.UserID)
	if queryErr != nil {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	if user.IsDeleted == 1 {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "账号或密码错误")
	}
	if user.Status != consts.StatusEnabled {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeForbidden, "账号已被禁用，请联系管理员")
	}
	// token_version 用于全量失效用户历史会话：改密码/改组/删除恢复等都会触发版本变化。
	if user.TokenVersion != claims.TokenVersion || user.TokenVersion != session.TokenVersion {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	// 普通用户需要保证所属用户组可用；超级管理员（group_id=0）不受用户组禁用影响。
	if user.GroupID != 0 {
		group, groupErr := c.GetGroupByID(ctx, user.GroupID)
		if groupErr != nil || group.Status != consts.StatusEnabled {
			return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeForbidden, "用户组已被禁用")
		}
	}
	return modelruntime.Principal{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: claims.ID}, user, nil
}
