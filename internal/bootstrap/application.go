package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	admincontroller "myjob/internal/controller/admin"
	"myjob/internal/kernel"
	authlib "myjob/internal/library/auth"
	smslib "myjob/internal/library/sms"
	adminlogic "myjob/internal/logic/admin"
	"myjob/internal/middleware"
	modelconfig "myjob/internal/model/config"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/redis/go-redis/v9"
)

type SMSConfig = modelruntime.SMSConfig

type Application struct {
	core   *kernel.Core
	server *ghttp.Server
}

func NewApplicationFromConfig(cfg modelconfig.Config) (*Application, error) {
	core, err := kernel.NewCoreFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return assemble(core), nil
}

func NewApplicationFromEnv() (*Application, error) {
	configPath := os.Getenv("ADMIN_CONFIG")
	if configPath == "" {
		// 运行时统一走本地真实开发配置，不再回退到 mock 示例模板。
		configPath = "manifest/config/config.local.yaml"
		if _, err := os.Stat(configPath); err != nil {
			return nil, err
		}
	}
	cfg, err := modelconfig.Load(configPath)
	if err != nil {
		return nil, err
	}
	return NewApplicationFromConfig(cfg)
}

func NewTestApplication() (*Application, error) {
	core, err := kernel.NewTestCore()
	if err != nil {
		return nil, err
	}
	app := assemble(core)
	// 测试里显式启动随机端口 server，确保 GoFrame 内部组件初始化完整。
	app.server.SetAddr("127.0.0.1:0")
	if err = app.server.Start(); err != nil {
		_ = core.Close()
		return nil, err
	}
	return app, nil
}

func assemble(core *kernel.Core) *Application {
	services := adminlogic.NewServices(core)
	authCtrl := admincontroller.NewAuth(services.Auth)
	userCtrl := admincontroller.NewUser(services.User)
	groupCtrl := admincontroller.NewGroup(services.Group)
	subjectCtrl := admincontroller.NewSubject(services.Subject)
	configCtrl := admincontroller.NewConfig(services.SMSConfig)
	logCtrl := admincontroller.NewLog(services.AuditLog)
	guard := middleware.NewAuthGuard(core)

	s := ghttp.GetServer(fmt.Sprintf("myjob-admin-%d", time.Now().UnixNano()))
	s.SetAddr(core.Config().Server.Address)
	// 路由绑定统一收口在这里，避免重新出现“各处散落注册”的旧 app 结构。
	s.BindHandler("POST:/api/admin/login", authCtrl.Login)
	s.BindHandler("POST:/api/admin/login/sms/send", authCtrl.LoginSMSSend)
	s.BindHandler("POST:/api/admin/login/sms/verify", authCtrl.LoginSMSVerify)
	s.BindHandler("POST:/api/admin/me", guard.Wrap("", false, authCtrl.Me))
	s.BindHandler("POST:/api/admin/logout", guard.Wrap("", false, authCtrl.Logout))

	s.BindHandler("GET:/api/admin/user/list", guard.Wrap("admin.list", false, userCtrl.List))
	s.BindHandler("POST:/api/admin/user/add", guard.Wrap("admin.list", false, userCtrl.Add))
	s.BindHandler("PUT:/api/admin/user/{id}", guard.Wrap("admin.list", false, userCtrl.Edit))
	s.BindHandler("DELETE:/api/admin/user/{id}", guard.Wrap("admin.list", false, userCtrl.Delete))
	s.BindHandler("PUT:/api/admin/user/{id}/status", guard.Wrap("admin.list", false, userCtrl.Status))
	s.BindHandler("PUT:/api/admin/user/{id}/notify", guard.Wrap("admin.list", false, userCtrl.Notify))
	s.BindHandler("POST:/api/admin/user/setBusiness", guard.Wrap("admin.list", false, userCtrl.SetBusiness))
	s.BindHandler("POST:/api/admin/user/cancelBusiness", guard.Wrap("admin.list", false, userCtrl.CancelBusiness))
	s.BindHandler("GET:/api/admin/user/trash", guard.Wrap("admin.list", false, userCtrl.Trash))
	s.BindHandler("PUT:/api/admin/user/{id}/restore", guard.Wrap("admin.list", false, userCtrl.Restore))

	s.BindHandler("GET:/api/admin/subject/list", guard.Wrap("subject.manage", false, subjectCtrl.List))
	s.BindHandler("POST:/api/admin/subject/add", guard.Wrap("subject.manage", false, subjectCtrl.Add))
	s.BindHandler("PUT:/api/admin/subject/{id}", guard.Wrap("subject.manage", false, subjectCtrl.Edit))

	s.BindHandler("GET:/api/admin/group/list", guard.Wrap("admin.department", false, groupCtrl.List))
	s.BindHandler("POST:/api/admin/group/add", guard.Wrap("admin.department", false, groupCtrl.Add))
	s.BindHandler("PUT:/api/admin/group/{id}", guard.Wrap("admin.department", false, groupCtrl.Edit))
	s.BindHandler("DELETE:/api/admin/group/{id}", guard.Wrap("admin.department", false, groupCtrl.Delete))
	s.BindHandler("PUT:/api/admin/group/{id}/status", guard.Wrap("admin.department", false, groupCtrl.Status))
	s.BindHandler("GET:/api/admin/group/{id}/auth", guard.Wrap("admin.department", false, groupCtrl.AuthGet))
	s.BindHandler("PUT:/api/admin/group/{id}/auth", guard.Wrap("admin.department", false, groupCtrl.AuthSave))
	s.BindHandler("GET:/api/admin/menu/tree", guard.Wrap("admin.department", false, groupCtrl.MenuTree))

	s.BindHandler("GET:/api/admin/config/sms", guard.Wrap("", true, configCtrl.GetSMS))
	s.BindHandler("PUT:/api/admin/config/sms", guard.Wrap("", true, configCtrl.SaveSMS))
	s.BindHandler("GET:/api/admin/log/operation", guard.Wrap("admin.action", false, logCtrl.Operation))
	s.BindHandler("GET:/api/admin/log/login", guard.Wrap("admin.loginlog", false, logCtrl.Login))

	return &Application{core: core, server: s}
}

func (a *Application) Handler() http.Handler { return a.server }
func (a *Application) Close() error          { _ = a.server.Shutdown(); return a.core.Close() }
func (a *Application) Redis() *redis.Client  { return a.core.Redis() }
func (a *Application) LastMockSMSCode(phone string) (string, error) {
	return a.core.LastMockSMSCode(phone)
}
func (a *Application) CreateTestUser(ctx context.Context, username, password, phone string) (int64, error) {
	return a.core.CreateTestUser(ctx, username, password, phone)
}
func (a *Application) SetSMSSender(sender smslib.Sender) { a.core.SetSMSSender(sender) }
func (a *Application) Core() *kernel.Core                { return a.core }
func (a *Application) Server() *ghttp.Server             { return a.server }
func SMSCodeKey(userID int64) string                     { return authlib.SMSCodeKey(userID) }
func SMSSendLockKey(userID int64) string                 { return authlib.SMSSendLockKey(userID) }
