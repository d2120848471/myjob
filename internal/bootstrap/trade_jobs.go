package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myjob/internal/app"
	tradelogic "myjob/internal/logic/trade"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/google/uuid"
)

// registerTradeJobs 在应用启动时注册 trade 定时任务。
//
// 注意：
// - 测试环境不自动启动（避免异步任务干扰测试用例）。
// - 多实例部署需要分布式锁，避免重复查单。
func registerTradeJobs(core *app.Core, trade *tradelogic.TradeOrderLogic) ([]string, error) {
	if core == nil || trade == nil {
		return nil, nil
	}
	if strings.TrimSpace(core.Config().AppEnv) == "test" {
		return nil, nil
	}

	name, err := registerTradeQueryJob(core, trade)
	if err != nil {
		return nil, err
	}
	return []string{name}, nil
}

func registerTradeQueryJob(core *app.Core, trade *tradelogic.TradeOrderLogic) (string, error) {
	interval := core.Config().Trade.AttemptQueryScanIntervalSeconds
	if interval <= 0 {
		interval = 30
	}
	pattern := fmt.Sprintf("@every %ds", interval)
	name := fmt.Sprintf("trade-query-job-%d", time.Now().UnixNano())

	_, err := gcron.AddSingleton(context.Background(), pattern, func(ctx context.Context) {
		ok, _ := tryAcquireRedisLock(ctx, core, "trade:query-job:lock", time.Duration(interval)*time.Second)
		if !ok {
			return
		}
		_, _ = trade.RunQueryJob(ctx, uuid.NewString())
	}, name)
	return name, err
}

func tryAcquireRedisLock(ctx context.Context, core *app.Core, key string, ttl time.Duration) (bool, error) {
	if core == nil || strings.TrimSpace(key) == "" {
		return false, nil
	}
	seconds := int64(ttl.Seconds())
	if seconds <= 0 {
		seconds = 10
	}
	value, err := core.Redis().GroupString().Set(ctx, key, "1", gredis.SetOption{
		TTLOption: gredis.TTLOption{EX: &seconds},
		NX:        true,
	})
	if err != nil {
		return false, err
	}
	if value == nil || value.IsNil() {
		return false, nil
	}
	return true, nil
}

