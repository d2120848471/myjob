package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myjob/internal/library/audit"
	"myjob/internal/library/region"
	"myjob/internal/library/sms"
	modelconfig "myjob/internal/model/config"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// NewCoreFromConfig 使用传入配置创建 Core，并完成 DB/Redis 初始化与启动引导（建表/种子数据）。
func NewCoreFromConfig(cfg modelconfig.Config) (*Core, error) {
	modelconfig.Normalize(&cfg)
	return newCore(cfg, nil, "")
}

// NewCoreFromConfigFile 从配置文件创建 Core。
//
// 配置缺省项（如超级管理员账号）会从环境变量补齐：SUPER_ADMIN_PHONE/SUPER_ADMIN_PASSWORD。
func NewCoreFromConfigFile(configPath string) (*Core, error) {
	cfgInstance, cfgName, cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return newCore(cfg, cfgInstance, cfgName)
}

// NewCoreFromEnv 从环境变量指定的配置文件创建 Core。
//
// - ADMIN_CONFIG 为空时默认使用 manifest/config/config.local.yaml
func NewCoreFromEnv() (*Core, error) {
	configPath := strings.TrimSpace(os.Getenv("ADMIN_CONFIG"))
	if configPath == "" {
		configPath = "manifest/config/config.local.yaml"
	}
	return NewCoreFromConfigFile(configPath)
}

// NewTestCore 构建用于测试的 Core：
// - 使用临时 sqlite 文件作为数据库
// - 使用 miniredis 作为 Redis
// - 使用 mock 短信发送器、同步审计写入、临时上传目录
//
// 调用 Close 会自动清理这些临时资源。
func NewTestCore() (*Core, error) {
	cfg := modelconfig.Default()
	cfg.AppEnv = "test"
	cfg.Database.Driver = "sqlite"
	cfg.Bootstrap.SuperAdminPhone = "13800000000"
	cfg.Bootstrap.SuperAdminPassword = "Admin_123"
	cfg.SMS.Provider = "mock"
	cfg.Audit.Async = false
	cfg.Upload.LocalDir = filepath.Join(os.TempDir(), fmt.Sprintf("myjob-upload-%d", time.Now().UnixNano()))
	cfg.Upload.PublicPrefix = "/uploads"
	cfg.Upload.MaxImageSizeMB = 2

	tmpFile, err := os.CreateTemp("", "myjob-admin-*.db")
	if err != nil {
		return nil, err
	}
	if err = tmpFile.Close(); err != nil {
		return nil, err
	}
	cfg.Database.DSN = tmpFile.Name()

	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	cfg.Redis.Addr = mr.Addr()
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0

	core, err := newCore(cfg, g.Cfg(fmt.Sprintf("myjob-test-%d", time.Now().UnixNano())), "")
	if err != nil {
		mr.Close()
		_ = os.Remove(tmpFile.Name())
		return nil, err
	}
	core.tempDBFile = tmpFile.Name()
	core.tempUploadDir = cfg.Upload.LocalDir
	core.miniRedis = mr
	return core, nil
}

// loadConfig 将指定配置文件加载到独立的 GoFrame 配置实例中，避免污染全局配置。
func loadConfig(configPath string) (*gcfg.Config, string, modelconfig.Config, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, "", modelconfig.Config{}, err
	}
	if _, err = os.Stat(absPath); err != nil {
		return nil, "", modelconfig.Config{}, err
	}
	cfgName := fmt.Sprintf("myjob-%d.yaml", time.Now().UnixNano())
	cfgInstance := g.Cfg(cfgName)
	adapter, ok := cfgInstance.GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		return nil, "", modelconfig.Config{}, errors.New("goframe config adapter is not file adapter")
	}
	adapter.SetFileName(absPath)
	cfg, err := modelconfig.LoadFromGoFrame(context.Background(), cfgInstance)
	if err != nil {
		return nil, "", modelconfig.Config{}, err
	}
	if cfg.Bootstrap.SuperAdminPhone == "" {
		cfg.Bootstrap.SuperAdminPhone = strings.TrimSpace(os.Getenv("SUPER_ADMIN_PHONE"))
	}
	if cfg.Bootstrap.SuperAdminPassword == "" {
		cfg.Bootstrap.SuperAdminPassword = strings.TrimSpace(os.Getenv("SUPER_ADMIN_PASSWORD"))
	}
	modelconfig.Normalize(&cfg)
	return cfgInstance, cfgName, cfg, nil
}

// newCore 构建 Core 并完成必要初始化：
// - 校验运行时必需配置（超级管理员账号）
// - 初始化 DB/Redis 连接与连通性检查
// - 启动引导（建表/补列/种子数据/默认配置）
// - 初始化审计写入器
func newCore(cfg modelconfig.Config, cfgInstance *gcfg.Config, cfgName string) (*Core, error) {
	if cfg.Bootstrap.SuperAdminPhone == "" {
		cfg.Bootstrap.SuperAdminPhone = strings.TrimSpace(os.Getenv("SUPER_ADMIN_PHONE"))
	}
	if cfg.Bootstrap.SuperAdminPassword == "" {
		cfg.Bootstrap.SuperAdminPassword = strings.TrimSpace(os.Getenv("SUPER_ADMIN_PASSWORD"))
	}
	if cfg.Bootstrap.SuperAdminPhone == "" || cfg.Bootstrap.SuperAdminPassword == "" {
		return nil, errors.New("SUPER_ADMIN_PHONE and SUPER_ADMIN_PASSWORD are required for runtime bootstrap")
	}
	core := &Core{
		cfg:            cfg,
		cfgName:        cfgName,
		cfgInstance:    cfgInstance,
		dbGroup:        fmt.Sprintf("myjob-db-%d", time.Now().UnixNano()),
		redisGroup:     fmt.Sprintf("myjob-redis-%d", time.Now().UnixNano()),
		driver:         cfg.Database.Driver,
		now:            time.Now,
		sender:         sms.NewSender(cfg.SMS.Provider),
		regionResolver: region.NewRegionResolver(region.DefaultDBPaths()...),
	}
	if mock, ok := core.sender.(*sms.MockSender); ok {
		core.mock = mock
	}
	if err := core.initStores(context.Background()); err != nil {
		return nil, err
	}
	if err := core.bootstrap(context.Background()); err != nil {
		_ = core.Close()
		return nil, err
	}
	core.initAuditWriter()
	return core, nil
}

// initStores 初始化数据库与 Redis 连接（使用独立 group），并进行连通性检查。
func (c *Core) initStores(ctx context.Context) error {
	dbNode := gdb.ConfigNode{
		Type:             c.cfg.Database.Driver,
		Link:             dbLink(c.cfg.Database.Driver, c.cfg.Database.DSN),
		MaxOpenConnCount: 10,
		MaxIdleConnCount: 5,
	}
	if c.cfg.Database.Driver == "sqlite" {
		dbNode.MaxOpenConnCount = 1
		dbNode.MaxIdleConnCount = 1
	}
	if err := gdb.SetConfigGroup(c.dbGroup, gdb.ConfigGroup{dbNode}); err != nil {
		return err
	}
	if err := c.DB().GetCore().PingMaster(); err != nil {
		return err
	}

	gredis.SetConfig(&gredis.Config{
		Address: c.cfg.Redis.Addr,
		Pass:    c.cfg.Redis.Password,
		Db:      c.cfg.Redis.DB,
	}, c.redisGroup)
	if _, err := c.Redis().Do(ctx, "PING"); err != nil {
		return err
	}
	return nil
}

// dbLink 根据 driver/dsn 组装 GoFrame 数据库连接字符串。
func dbLink(driver, dsn string) string {
	driver = strings.TrimSpace(strings.ToLower(driver))
	dsn = strings.TrimSpace(dsn)
	switch driver {
	case "sqlite":
		if strings.HasPrefix(dsn, "sqlite::") {
			return dsn
		}
		return fmt.Sprintf("sqlite::@file(%s)", dsn)
	case "mysql":
		if strings.HasPrefix(dsn, "mysql:") {
			return dsn
		}
		return "mysql:" + dsn
	default:
		return dsn
	}
}

// initAuditWriter 初始化操作审计写入器（同步或异步由配置决定）。
func (c *Core) initAuditWriter() {
	c.auditWriter = audit.NewWriter(!c.cfg.Audit.Async, c.cfg.Audit.BufferSize, c.insertOperationLog)
	c.auditWriter.Start()
}

// Close 释放 Core 持有的资源（审计写入器、Redis/DB 连接、测试临时资源等）。
func (c *Core) Close() error {
	if c.auditWriter != nil {
		c.auditWriter.Close()
	}
	if c.Redis() != nil {
		_ = c.Redis().Close(context.Background())
	}
	var err error
	if c.DB() != nil {
		err = c.DB().Close(context.Background())
	}
	if c.miniRedis != nil {
		c.miniRedis.Close()
	}
	if c.tempDBFile != "" {
		_ = os.Remove(c.tempDBFile)
	}
	if c.tempUploadDir != "" {
		_ = os.RemoveAll(c.tempUploadDir)
	}
	return err
}
