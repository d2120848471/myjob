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

func (c *Core) AuthenticateRequest(ctx context.Context, authz string) (modelruntime.Principal, AdminUser, error) {
	authz = strings.TrimSpace(authz)
	if !strings.HasPrefix(authz, "Bearer ") {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	claims, err := authlib.ParseToken(c.cfg.Auth.JWTSecret, tokenString)
	if err != nil || claims.ID == "" {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
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
	if user.TokenVersion != claims.TokenVersion || user.TokenVersion != session.TokenVersion {
		return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	if user.GroupID != 0 {
		group, groupErr := c.GetGroupByID(ctx, user.GroupID)
		if groupErr != nil || group.Status != consts.StatusEnabled {
			return modelruntime.Principal{}, AdminUser{}, gerror.NewCode(consts.CodeForbidden, "用户组已被禁用")
		}
	}
	return modelruntime.Principal{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: claims.ID}, user, nil
}
