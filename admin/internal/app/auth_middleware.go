package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authedHandler func(http.ResponseWriter, *http.Request, principal, AdminUser)

func (a *Application) withAuth(permission string, superOnly bool, next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, user, err := a.authenticateRequest(r.Context(), r)
		if err != nil {
			writeError(w, err.HTTPStatus, err.Code, err.Message)
			return
		}
		if superOnly && user.GroupID != 0 {
			writeError(w, http.StatusForbidden, 403, "仅超级管理员可访问")
			return
		}
		if permission != "" && user.GroupID != 0 {
			perms, loadErr := a.loadPermissions(r.Context(), user.GroupID)
			if loadErr != nil {
				writeError(w, http.StatusInternalServerError, 500, "权限加载失败")
				return
			}
			if !containsString(perms, permission) {
				writeError(w, http.StatusForbidden, 403, "无权限访问")
				return
			}
		}
		next(w, r, p, user)
	}
}

func (a *Application) authenticateRequest(ctx context.Context, r *http.Request) (principal, AdminUser, *apiError) {
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authz, "Bearer ") {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	if claims.ID == "" {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	var session sessionPayload
	raw, redisErr := a.redis.Get(ctx, sessionKey(claims.ID)).Result()
	if redisErr != nil {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	if err = json.Unmarshal([]byte(raw), &session); err != nil {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	user, queryErr := a.getUserByID(ctx, claims.UserID)
	if queryErr != nil {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	if user.IsDeleted == 1 {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "账号或密码错误"}
	}
	if user.Status != statusEnabled {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusForbidden, Code: 403, Message: "账号已被禁用，请联系管理员"}
	}
	if user.TokenVersion != claims.TokenVersion || user.TokenVersion != session.TokenVersion {
		return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusUnauthorized, Code: 401, Message: "未登录或登录已失效"}
	}
	if user.GroupID != 0 {
		group, groupErr := a.getGroupByID(ctx, user.GroupID)
		if groupErr != nil || group.Status != statusEnabled {
			return principal{}, AdminUser{}, &apiError{HTTPStatus: http.StatusForbidden, Code: 403, Message: "用户组已被禁用"}
		}
	}
	return principal{UserID: user.ID, GroupID: user.GroupID, TokenVersion: user.TokenVersion, JTI: claims.ID}, user, nil
}
