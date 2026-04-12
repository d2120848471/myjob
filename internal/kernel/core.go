package kernel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"myjob/internal/library/audit"
	"myjob/internal/library/region"
	"myjob/internal/library/sms"
	"myjob/internal/model/config"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/redis/go-redis/v9"
)

func NewCoreFromConfig(cfg config.Config) (*Core, error) {
	if cfg.Bootstrap.SuperAdminPhone == "" {
		cfg.Bootstrap.SuperAdminPhone = os.Getenv("SUPER_ADMIN_PHONE")
	}
	if cfg.Bootstrap.SuperAdminPassword == "" {
		cfg.Bootstrap.SuperAdminPassword = os.Getenv("SUPER_ADMIN_PASSWORD")
	}
	if cfg.Bootstrap.SuperAdminPhone == "" || cfg.Bootstrap.SuperAdminPassword == "" {
		return nil, errors.New("SUPER_ADMIN_PHONE and SUPER_ADMIN_PASSWORD are required for runtime bootstrap")
	}
	db, err := gdb.New(gdb.ConfigNode{Type: cfg.Database.Driver, Link: dbLink(cfg.Database.Driver, cfg.Database.DSN), MaxOpenConnCount: 10, MaxIdleConnCount: 5})
	if err != nil {
		return nil, err
	}
	if err = db.GetCore().PingMaster(); err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	core := &Core{
		cfg:            cfg,
		driver:         cfg.Database.Driver,
		db:             db,
		redis:          rdb,
		now:            time.Now,
		sender:         sms.NewSender(cfg.SMS.Provider),
		regionResolver: region.NewRegionResolver(region.DefaultDBPaths()...),
	}
	if mock, ok := core.sender.(*sms.MockSender); ok {
		core.mock = mock
	}
	if err = core.bootstrap(context.Background()); err != nil {
		return nil, err
	}
	core.initAuditWriter()
	return core, nil
}

func NewTestCore() (*Core, error) {
	cfg := config.Default()
	cfg.AppEnv = "test"
	cfg.Database.Driver = "sqlite"
	cfg.Bootstrap.SuperAdminPhone = "13800000000"
	cfg.Bootstrap.SuperAdminPassword = "Admin_123"
	cfg.SMS.Provider = "mock"
	cfg.Audit.Async = false
	tmpFile, err := os.CreateTemp("", "myjob-admin-*.db")
	if err != nil {
		return nil, err
	}
	_ = tmpFile.Close()
	cfg.Database.DSN = tmpFile.Name()

	db, err := gdb.New(gdb.ConfigNode{Type: "sqlite", Link: dbLink("sqlite", cfg.Database.DSN), MaxOpenConnCount: 1, MaxIdleConnCount: 1})
	if err != nil {
		return nil, err
	}
	if err = db.GetCore().PingMaster(); err != nil {
		return nil, err
	}
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	core := &Core{
		cfg:            cfg,
		driver:         "sqlite",
		db:             db,
		redis:          rdb,
		now:            time.Now,
		sender:         sms.NewSender(cfg.SMS.Provider),
		regionResolver: region.NewRegionResolver(region.DefaultDBPaths()...),
	}
	core.mock = core.sender.(*sms.MockSender)
	if err = core.bootstrap(context.Background()); err != nil {
		return nil, err
	}
	core.initAuditWriter()
	return core, nil
}

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

func (c *Core) initAuditWriter() {
	c.auditWriter = audit.NewWriter(!c.cfg.Audit.Async, c.cfg.Audit.BufferSize, c.insertOperationLog)
	c.auditWriter.Start()
}

func (c *Core) Close() error {
	if c.auditWriter != nil {
		c.auditWriter.Close()
	}
	if c.redis != nil {
		_ = c.redis.Close()
	}
	var err error
	if c.db != nil {
		err = c.db.Close(context.Background())
	}
	if c.tempDBFile != "" {
		_ = os.Remove(c.tempDBFile)
	}
	return err
}
