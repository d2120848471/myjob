package adminlogic

import (
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func apiErr(code gcode.Code, message string) error {
	return gerror.NewCode(code, message)
}

// Services 聚合后台各业务模块的 service 实现（由 logic 层提供具体实现）。
type Services struct {
	Auth                       service.AuthService
	User                       service.UserService
	Group                      service.GroupService
	Subject                    service.SubjectService
	Brand                      service.BrandService
	Industry                   service.IndustryService
	ProductTemplate            service.ProductTemplateService
	PurchaseLimit              service.PurchaseLimitService
	ProductGoods               service.ProductGoodsService
	ProductGoodsChannelConfig  service.ProductGoodsChannelConfigService
	ProductGoodsChannelBinding service.ProductGoodsChannelBindingService
	SupplierPlatform           service.SupplierPlatformService
	SMSConfig                  service.SMSConfigService
	System                     service.SystemConfigService
	AuditLog                   service.AuditLogService
}

// NewServices 基于 core 构建一组后台服务实现，供 controller 注入使用。
func NewServices(core *app.Core) *Services {
	return &Services{
		Auth:                       &AuthLogic{core: core},
		User:                       &UserLogic{core: core},
		Group:                      &GroupLogic{core: core},
		Subject:                    &SubjectLogic{core: core},
		Brand:                      &BrandLogic{core: core},
		Industry:                   &IndustryLogic{core: core},
		ProductTemplate:            &ProductTemplateLogic{core: core},
		PurchaseLimit:              &PurchaseLimitLogic{core: core},
		ProductGoods:               &ProductGoodsLogic{core: core},
		ProductGoodsChannelConfig:  &ProductGoodsChannelConfigLogic{core: core},
		ProductGoodsChannelBinding: &ProductGoodsChannelBindingLogic{core: core},
		SupplierPlatform:           NewSupplierPlatformLogic(core),
		SMSConfig:                  &SMSConfigLogic{core: core},
		System:                     &SystemConfigLogic{core: core},
		AuditLog:                   &AuditLogLogic{core: core},
	}
}

var _ = consts.CodeInternalError
