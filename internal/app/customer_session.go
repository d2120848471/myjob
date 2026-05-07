package app

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"myjob/internal/consts"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomerSessionKey 返回单个客户会话 Redis key。
func CustomerSessionKey(jti string) string { return "customer:session:" + jti }

// CustomerSessionsKey 返回客户活跃会话集合 Redis key。
func CustomerSessionsKey(customerID int64) string {
	return fmt.Sprintf("customer:user:sessions:%d", customerID)
}

type customerClaims struct {
	CustomerID   int64 `json:"customer_id"`
	TokenVersion int   `json:"token_version"`
	jwt.RegisteredClaims
}

// BuildCustomerLoginUser 构建客户登录返回视图。
func (c *Core) BuildCustomerLoginUser(customer CustomerUser) modelruntime.CustomerLoginUser {
	return modelruntime.CustomerLoginUser{
		ID:          customer.ID,
		CompanyName: customer.CompanyName,
		Phone:       customer.Phone,
		Status:      customer.Status,
	}
}

// IssueCustomerSession 签发客户 token 并保存服务端 session。
func (c *Core) IssueCustomerSession(ctx context.Context, customer CustomerUser) (string, error) {
	ttl := time.Duration(c.cfg.Auth.AccessTokenTTLMin) * time.Minute
	now := c.now()
	jti := uuid.NewString()
	expiresAt := now.Add(ttl)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customerClaims{
		CustomerID:   customer.ID,
		TokenVersion: customer.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	})
	tokenString, err := token.SignedString([]byte(c.cfg.Auth.JWTSecret))
	if err != nil {
		return "", err
	}
	payload := modelruntime.CustomerSessionPayload{
		CustomerID:   customer.ID,
		TokenVersion: customer.TokenVersion,
		JTI:          jti,
		ExpiresAt:    expiresAt,
	}
	if err = saveCustomerSession(ctx, c.Redis(), payload, ttl); err != nil {
		return "", err
	}
	return tokenString, nil
}

func saveCustomerSession(ctx context.Context, client *gredis.Redis, payload modelruntime.CustomerSessionPayload, ttl time.Duration) error {
	data, _ := json.Marshal(payload)
	seconds := int64(math.Ceil(ttl.Seconds()))
	if _, err := client.GroupString().Set(ctx, CustomerSessionKey(payload.JTI), string(data), gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
	}); err != nil {
		return err
	}
	if _, err := client.GroupSet().SAdd(ctx, CustomerSessionsKey(payload.CustomerID), payload.JTI); err != nil {
		return err
	}
	_, _ = client.GroupGeneric().Expire(ctx, CustomerSessionsKey(payload.CustomerID), seconds)
	return nil
}

// RemoveAllCustomerSessions 删除客户所有活跃 session，用于改登录密码、禁用和删除。
func (c *Core) RemoveAllCustomerSessions(ctx context.Context, customerID int64) error {
	sessions, err := c.RedisSMembers(ctx, CustomerSessionsKey(customerID))
	if err == nil {
		for _, jti := range sessions {
			_, _ = c.Redis().GroupGeneric().Del(ctx, CustomerSessionKey(jti))
		}
	}
	_, err = c.Redis().GroupGeneric().Del(ctx, CustomerSessionsKey(customerID))
	return err
}

// AuthenticateCustomerRequest 校验客户 Bearer token；V1 预留给后续客户中心/下单鉴权使用。
func (c *Core) AuthenticateCustomerRequest(ctx context.Context, authz string) (modelruntime.CustomerPrincipal, CustomerUser, error) {
	authz = strings.TrimSpace(authz)
	if !strings.HasPrefix(authz, "Bearer ") {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	claims := &customerClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(c.cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid || claims.ID == "" {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	var session modelruntime.CustomerSessionPayload
	raw, err := c.RedisGetString(ctx, CustomerSessionKey(claims.ID))
	if err != nil {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	if err = json.Unmarshal([]byte(raw), &session); err != nil {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	customer, err := c.GetCustomerByID(ctx, claims.CustomerID)
	if err != nil || customer.IsDeleted == 1 {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	if customer.Status != consts.StatusEnabled {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeForbidden, "账号已被禁用，请联系客服")
	}
	if customer.TokenVersion != claims.TokenVersion || customer.TokenVersion != session.TokenVersion {
		return modelruntime.CustomerPrincipal{}, CustomerUser{}, gerror.NewCode(consts.CodeUnauthorized, "未登录或登录已失效")
	}
	return modelruntime.CustomerPrincipal{CustomerID: customer.ID, TokenVersion: customer.TokenVersion, JTI: claims.ID}, customer, nil
}

// GetCustomerByID 根据 ID 查询客户账号。
func (c *Core) GetCustomerByID(ctx context.Context, id int64) (CustomerUser, error) {
	var customer CustomerUser
	err := c.DB().GetCore().GetScan(ctx, &customer, `
SELECT id, company_name, phone, password_hash, pay_password_hash, status, is_deleted,
       last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM customer_user
WHERE id = ?
`, id)
	return customer, err
}

// GetCustomerByPhone 根据手机号查询客户账号，包括回收站客户。
func (c *Core) GetCustomerByPhone(ctx context.Context, phone string) (CustomerUser, error) {
	var customer CustomerUser
	err := c.DB().GetCore().GetScan(ctx, &customer, `
SELECT id, company_name, phone, password_hash, pay_password_hash, status, is_deleted,
       last_login_ip, last_login_at, token_version, deleted_at, created_at, updated_at
FROM customer_user
WHERE phone = ?
`, phone)
	return customer, err
}
