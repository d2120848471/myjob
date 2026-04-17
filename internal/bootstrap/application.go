package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"myjob/internal/app"
	admincontroller "myjob/internal/controller/admin"
	opencontroller "myjob/internal/controller/open"
	providercontroller "myjob/internal/controller/provider"
	authlib "myjob/internal/library/auth"
	smslib "myjob/internal/library/sms"
	adminlogic "myjob/internal/logic/admin"
	tradelogic "myjob/internal/logic/trade"
	"myjob/internal/middleware"
	modelconfig "myjob/internal/model/config"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcron"
)

// SMSConfig 是运行态短信配置快照的别名，便于测试与对外透出。
type SMSConfig = modelruntime.SMSConfig

// Application 表示 MyJob 后台应用实例，持有运行时 Core 与 HTTP Server，并提供测试辅助能力。
type Application struct {
	core      *app.Core
	server    *ghttp.Server
	cronNames []string
}

// NewApplicationFromConfig 基于显式配置创建应用实例（不会自动启动 HTTP Server）。
func NewApplicationFromConfig(cfg modelconfig.Config) (*Application, error) {
	core, err := app.NewCoreFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	appInstance, err := assemble(core)
	if err != nil {
		_ = core.Close()
		return nil, err
	}
	return appInstance, nil
}

// NewApplicationFromEnv 基于环境变量配置创建应用实例（不会自动启动 HTTP Server）。
func NewApplicationFromEnv() (*Application, error) {
	core, err := app.NewCoreFromEnv()
	if err != nil {
		return nil, err
	}
	appInstance, err := assemble(core)
	if err != nil {
		_ = core.Close()
		return nil, err
	}
	return appInstance, nil
}

// NewTestApplication 创建用于集成测试的应用实例：使用测试 Core，绑定随机端口并启动 HTTP Server。
func NewTestApplication() (*Application, error) {
	core, err := app.NewTestCore()
	if err != nil {
		return nil, err
	}
	appInstance, err := assemble(core)
	if err != nil {
		_ = core.Close()
		return nil, err
	}
	appInstance.server.SetAddr("127.0.0.1:0")
	if err = appInstance.server.Start(); err != nil {
		_ = core.Close()
		return nil, err
	}
	return appInstance, nil
}

func assemble(core *app.Core) (*Application, error) {
	services := adminlogic.NewServices(core)
	authCtrl := admincontroller.NewAuth(services.Auth)
	sessionCtrl := admincontroller.NewSession(services.Auth)
	userCtrl := admincontroller.NewUser(services.User)
	groupCtrl := admincontroller.NewGroup(services.Group)
	subjectCtrl := admincontroller.NewSubject(services.Subject)
	brandCtrl := admincontroller.NewBrand(services.Brand)
	industryCtrl := admincontroller.NewIndustry(services.Industry)
	productTemplateCtrl := admincontroller.NewProductTemplate(services.ProductTemplate)
	purchaseLimitCtrl := admincontroller.NewPurchaseLimit(services.PurchaseLimit)
	productGoodsCtrl := admincontroller.NewProductGoods(services.ProductGoods)
	productGoodsChannelConfigCtrl := admincontroller.NewProductGoodsChannelConfig(services.ProductGoodsChannelConfig)
	productGoodsChannelBindingCtrl := admincontroller.NewProductGoodsChannelBinding(services.ProductGoodsChannelBinding)
	supplierPlatformCtrl := admincontroller.NewSupplierPlatform(services.SupplierPlatform)
	settingsCtrl := admincontroller.NewSettings(services.SMSConfig, services.System)
	operationLogCtrl := admincontroller.NewOperationLog(services.AuditLog)
	loginLogCtrl := admincontroller.NewLoginLog(services.AuditLog)
	guard := middleware.NewAuthGuard(core)
	openSignatureGuard := middleware.NewOpenSignatureGuard(core)
	openOrderCtrl := opencontroller.NewOpenOrder(tradelogic.NewOpenOrderLogic(core))
	tradeOrderLogic := tradelogic.NewTradeOrderLogic(core, nil, nil)
	providerCallbackCtrl := providercontroller.NewProviderCallback(tradeOrderLogic)
	providerPriceNotifyCtrl := providercontroller.NewProviderPriceNotify(tradeOrderLogic)

	s := ghttp.GetServer(fmt.Sprintf("myjob-admin-%d", time.Now().UnixNano()))
	s.SetAddr(core.Config().Server.Address)
	s.SetOpenApiPath("/api.json")
	s.SetSwaggerPath("/swagger")
	if core.Config().AppEnv == "test" {
		s.SetDumpRouterMap(false)
	}
	configureOpenAPI(s)
	if err := mountUploadStaticPath(s, core.Config().Upload); err != nil {
		return nil, err
	}

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
			group.Middleware(guard.Require("product.brand", false))
			group.Bind(brandCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("product.industry", false))
			group.Bind(industryCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("product.template", false))
			group.Bind(productTemplateCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("product.purchase_limit", false))
			group.Bind(purchaseLimitCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("product.goods", false))
			group.Bind(productGoodsCtrl)
			group.Bind(productGoodsChannelConfigCtrl)
			group.Bind(productGoodsChannelBindingCtrl)
		})
		group.Group("", func(group *ghttp.RouterGroup) {
			group.Middleware(guard.Require("supplier.index", false))
			group.Bind(supplierPlatformCtrl)
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

	s.Group("/api/open", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.Response)
		group.Middleware(openSignatureGuard.Require())
		group.Bind(openOrderCtrl)
	})

	s.Group("/api/provider", func(group *ghttp.RouterGroup) {
		group.Bind(providerCallbackCtrl)
		group.Bind(providerPriceNotifyCtrl)
	})

	cronNames, err := registerTradeJobs(core, tradeOrderLogic)
	if err != nil {
		return nil, err
	}
	return &Application{core: core, server: s, cronNames: cronNames}, nil
}

func mountUploadStaticPath(s *ghttp.Server, cfg modelconfig.UploadConfig) error {
	if strings.TrimSpace(cfg.PublicPrefix) == "" {
		return nil
	}
	if err := os.MkdirAll(cfg.LocalDir, 0o755); err != nil {
		return fmt.Errorf("初始化上传目录失败: %w", err)
	}
	s.AddStaticPath(cfg.PublicPrefix, cfg.LocalDir)
	return nil
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
func (a *Application) Close() error {
	_ = a.server.Shutdown()
	for _, name := range a.cronNames {
		gcron.Remove(name)
	}
	return a.core.Close()
}
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
