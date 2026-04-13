package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID       int64 `json:"user_id"`
	GroupID      int64 `json:"group_id"`
	TokenVersion int   `json:"token_version"`
	jwt.RegisteredClaims
}

func IssueToken(secret string, payload modelruntime.SessionPayload, now time.Time, ttl time.Duration) (string, string, error) {
	jti := uuid.NewString()
	expiresAt := now.Add(ttl)
	claims := Claims{
		UserID:       payload.UserID,
		GroupID:      payload.GroupID,
		TokenVersion: payload.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	return tokenString, jti, nil
}

func ParseToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func SaveSession(ctx context.Context, client *gredis.Redis, payload modelruntime.SessionPayload, ttl time.Duration) error {
	data, _ := json.Marshal(payload)
	seconds := int64(math.Ceil(ttl.Seconds()))
	if _, err := client.GroupString().Set(ctx, SessionKey(payload.JTI), string(data), gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
	}); err != nil {
		return err
	}
	if _, err := client.GroupSet().SAdd(ctx, UserSessionsKey(payload.UserID), payload.JTI); err != nil {
		return err
	}
	_, _ = client.GroupGeneric().Expire(ctx, UserSessionsKey(payload.UserID), seconds)
	return nil
}

func SessionKey(jti string) string {
	return "admin:session:" + jti
}

func UserSessionsKey(userID int64) string {
	return fmt.Sprintf("admin:user:sessions:%d", userID)
}

func TempLoginKey(loginToken string) string {
	return "admin:login:tmp:" + loginToken
}

func SMSCodeKey(userID int64) string {
	return fmt.Sprintf("sms:login:%d", userID)
}

func SMSSendLockKey(userID int64) string {
	return fmt.Sprintf("sms:login:send_lock:%d", userID)
}

func PermissionCacheKey(groupID int64) string {
	return fmt.Sprintf("admin:perm:group:%d", groupID)
}

func SMSConfigCacheKey() string {
	return "admin:config:sms"
}

func SystemConfigCacheKey(group string) string {
	return fmt.Sprintf("admin:config:system:%s", group)
}
