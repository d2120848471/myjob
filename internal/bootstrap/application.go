package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"myjob/internal/app"
	admincontroller "myjob/internal/controller/admin"
	authlib "myjob/internal/library/auth"
	smslib "myjob/internal/library/sms"
	adminlogic "myjob/internal/logic/admin"
	"myjob/internal/middleware"
	modelconfig "myjob/internal/model/config"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
)

type SMSConfig = modelruntime.SMSConfig

type Application struct {
	core   *app.Core
	server *ghttp.Server
}

func NewApplicationFromConfig(cfg modelconfig.Config) (*Application, error) {
	core, err := app.NewCoreFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	return assemble(core), nil
}

func NewApplicationFromEnv() (*Application, error) {
	core, err := app.NewCoreFromEnv()
	if err != nil {
		return nil, err
	}
	return assemble(core), nil
}

func NewTestApplication() (*Application, error) {
	core, err := app.NewTestCore()
	if err != nil {
		return nil, err
	}
	appInstance := assemble(core)
	appInstance.server.SetAddr("127.0.0.1:0")
	if err = appInstance.server.Start(); err != nil {
		_ = core.Close()
		return nil, err
	}
	return appInstance, nil
}

func assemble(core *app.Core) *Application {
	services := adminlogic.NewServices(core)
	authCtrl := admincontroller.NewAuth(services.Auth)
	sessionCtrl := admincontroller.NewSession(services.Auth)
	userCtrl := admincontroller.NewUser(services.User)
	groupCtrl := admincontroller.NewGroup(services.Group)
	subjectCtrl := admincontroller.NewSubject(services.Subject)
	settingsCtrl := admincontroller.NewSettings(services.SMSConfig)
	operationLogCtrl := admincontroller.NewOperationLog(services.AuditLog)
	loginLogCtrl := admincontroller.NewLoginLog(services.AuditLog)
	guard := middleware.NewAuthGuard(core)

	s := ghttp.GetServer(fmt.Sprintf("myjob-admin-%d", time.Now().UnixNano()))
	s.SetAddr(core.Config().Server.Address)
	s.SetOpenApiPath("/api.json")
	s.SetSwaggerPath("/swagger")
	configureOpenAPI(s)

	s.Group("/api/admin", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.Response)
		group.Bind(authCtrl)

		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("", false))
			group.Bind(sessionCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("admin.list", false))
			group.Bind(userCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("admin.department", false))
			group.Bind(groupCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("subject.manage", false))
			group.Bind(subjectCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("", true))
			group.Bind(settingsCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("admin.action", false))
			group.Bind(operationLogCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("admin.loginlog", false))
			group.Bind(loginLogCtrl)
		})
	})

	return &Application{core: core, server: s}
}

func configureOpenAPI(s *ghttp.Server) {
	oai := s.GetOpenApi()
	oai.Config.CommonResponse = middleware.JSONResponse{}
	oai.Config.CommonResponseDataField = "Data"
	oai.Components = goai.Components{
		SecuritySchemes: goai.SecuritySchemes{
			"BearerAuth": {
				Value: &goai.SecurityScheme{
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
		},
	}
	oai.Info.Title = "MyJob Admin API"
	oai.Info.Version = "v1"
}

func (a *Application) Handler() http.Handler { return a.server }
func (a *Application) Close() error          { _ = a.server.Shutdown(); return a.core.Close() }
func (a *Application) Redis() *gredis.Redis  { return a.core.Redis() }
func (a *Application) LastMockSMSCode(phone string) (string, error) {
	return a.core.LastMockSMSCode(phone)
}
func (a *Application) CreateTestUser(ctx context.Context, username, password, phone string) (int64, error) {
	return a.core.CreateTestUser(ctx, username, password, phone)
}
func (a *Application) SetSMSSender(sender smslib.Sender) { a.core.SetSMSSender(sender) }
func (a *Application) Core() *app.Core                   { return a.core }
func (a *Application) Server() *ghttp.Server             { return a.server }
func SMSCodeKey(userID int64) string                     { return authlib.SMSCodeKey(userID) }
func SMSSendLockKey(userID int64) string                 { return authlib.SMSSendLockKey(userID) }
