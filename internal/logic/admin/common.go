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

type Services struct {
	Auth      service.AuthService
	User      service.UserService
	Group     service.GroupService
	Subject   service.SubjectService
	Brand     service.BrandService
	Industry  service.IndustryService
	SMSConfig service.SMSConfigService
	System    service.SystemConfigService
	AuditLog  service.AuditLogService
}

func NewServices(core *app.Core) *Services {
	return &Services{
		Auth:      &AuthLogic{core: core},
		User:      &UserLogic{core: core},
		Group:     &GroupLogic{core: core},
		Subject:   &SubjectLogic{core: core},
		Brand:     &BrandLogic{core: core},
		Industry:  &IndustryLogic{core: core},
		SMSConfig: &SMSConfigLogic{core: core},
		System:    &SystemConfigLogic{core: core},
		AuditLog:  &AuditLogLogic{core: core},
	}
}

var _ = consts.CodeInternalError
