package app

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/gogf/gf/v2/database/gredis"
)

// RedisSetString 将字符串写入 Redis，并设置过期时间。
//
// ttl <= 0 时会强制设置为 1s，避免出现永久 key。
func (c *Core) RedisSetString(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = time.Second
	}
	seconds := int64(math.Ceil(ttl.Seconds()))
	_, err := c.Redis().GroupString().Set(ctx, key, value, gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
	})
	return err
}

// RedisGetString 从 Redis 读取字符串；当 key 不存在时返回 sql.ErrNoRows。
func (c *Core) RedisGetString(ctx context.Context, key string) (string, error) {
	value, err := c.Redis().GroupString().Get(ctx, key)
	if err != nil {
		return "", err
	}
	if value == nil || value.IsNil() {
		return "", sql.ErrNoRows
	}
	return value.String(), nil
}

// RedisTTL 读取 Redis key 的剩余 TTL。
func (c *Core) RedisTTL(ctx context.Context, key string) (time.Duration, error) {
	seconds, err := c.Redis().GroupGeneric().TTL(ctx, key)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}

// RedisSMembers 读取 Redis set 中的所有成员并转换为字符串切片。
func (c *Core) RedisSMembers(ctx context.Context, key string) ([]string, error) {
	values, err := c.Redis().GroupSet().SMembers(ctx, key)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result, nil
}
