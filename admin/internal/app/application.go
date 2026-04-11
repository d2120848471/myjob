package app

import (
	"admin/utility/ipx"
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func NewApplicationFromConfig(cfg Config) (*Application, error) {
	if cfg.Bootstrap.SuperAdminPhone == "" {
		cfg.Bootstrap.SuperAdminPhone = os.Getenv("SUPER_ADMIN_PHONE")
	}
	if cfg.Bootstrap.SuperAdminPassword == "" {
		cfg.Bootstrap.SuperAdminPassword = os.Getenv("SUPER_ADMIN_PASSWORD")
	}
	if cfg.Bootstrap.SuperAdminPhone == "" || cfg.Bootstrap.SuperAdminPassword == "" {
		return nil, errors.New("SUPER_ADMIN_PHONE and SUPER_ADMIN_PASSWORD are required for runtime bootstrap")
	}
	db, err := sqlx.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	app := &Application{
		cfg:            cfg,
		driver:         cfg.Database.Driver,
		db:             db,
		redis:          rdb,
		mux:            http.NewServeMux(),
		now:            time.Now,
		sender:         newMockSMSSender(),
		syncAudit:      !cfg.Audit.Async,
		auditCh:        make(chan operationEvent, cfg.Audit.BufferSize),
		auditStop:      make(chan struct{}),
		regionResolver: ipx.NewRegionResolver(ipx.DefaultDBPaths()...),
	}
	if sender, ok := app.sender.(*MockSMSSender); ok {
		app.mock = sender
	}
	if err = app.bootstrap(context.Background()); err != nil {
		return nil, err
	}
	app.startAuditWorker()
	app.registerRoutes()
	return app, nil
}

func NewTestApplication() (*Application, error) {
	cfg := defaultConfig()
	cfg.AppEnv = "test"
	cfg.Database.Driver = "sqlite"
	cfg.Database.DSN = ":memory:"
	cfg.Bootstrap.SuperAdminPhone = "13800000000"
	cfg.Bootstrap.SuperAdminPassword = "Admin_123"
	cfg.SMS.Provider = "mock"
	cfg.Audit.Async = false

	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	app := &Application{
		cfg:            cfg,
		driver:         "sqlite",
		db:             db,
		redis:          rdb,
		mux:            http.NewServeMux(),
		now:            time.Now,
		sender:         newMockSMSSender(),
		syncAudit:      true,
		auditCh:        make(chan operationEvent, 8),
		auditStop:      make(chan struct{}),
		regionResolver: ipx.NewRegionResolver(ipx.DefaultDBPaths()...),
	}
	app.mock = app.sender.(*MockSMSSender)
	if err = app.bootstrap(context.Background()); err != nil {
		return nil, err
	}
	app.registerRoutes()
	return app, nil
}

func (a *Application) Handler() http.Handler {
	return a.mux
}

func (a *Application) Close() error {
	close(a.auditStop)
	a.auditWG.Wait()
	if a.redis != nil {
		_ = a.redis.Close()
	}
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *Application) LastMockSMSCode(phone string) (string, error) {
	if a.mock == nil {
		return "", errors.New("mock sender not configured")
	}
	return a.mock.LastCode(phone)
}

func (a *Application) CreateTestUser(ctx context.Context, username, password, phone string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	result, err := a.db.ExecContext(ctx, `
INSERT INTO admin_user (
    username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, ?)
`, username, string(hash), username, phone, 1, statusEnabled, a.now(), a.now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (a *Application) bootstrap(ctx context.Context) error {
	if a.driver == "sqlite" {
		if err := execStatements(ctx, a.db, sqliteSchema); err != nil {
			return err
		}
	} else {
		if err := execStatements(ctx, a.db, mysqlSchema); err != nil {
			return err
		}
	}
	if err := a.ensureMenuSchema(ctx); err != nil {
		return err
	}
	if err := a.ensureDefaultGroup(ctx); err != nil {
		return err
	}
	if err := a.ensureMenus(ctx); err != nil {
		return err
	}
	if err := a.ensureDefaultGroupAuth(ctx); err != nil {
		return err
	}
	if err := a.ensureSMSConfig(ctx); err != nil {
		return err
	}
	if err := a.ensureSuperAdmin(ctx); err != nil {
		return err
	}
	return nil
}
