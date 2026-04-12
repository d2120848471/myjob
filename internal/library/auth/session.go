package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	modelruntime "myjob/internal/model/runtime"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func SaveSession(ctx context.Context, client *redis.Client, payload modelruntime.SessionPayload, ttl time.Duration) error {
	data, _ := json.Marshal(payload)
	if err := client.Set(ctx, SessionKey(payload.JTI), data, ttl).Err(); err != nil {
		return err
	}
	if err := client.SAdd(ctx, UserSessionsKey(payload.UserID), payload.JTI).Err(); err != nil {
		return err
	}
	_ = client.Expire(ctx, UserSessionsKey(payload.UserID), ttl).Err()
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
