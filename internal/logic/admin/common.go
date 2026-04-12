package adminlogic

import (
	"net/http"

	"myjob/internal/kernel"
	modelruntime "myjob/internal/model/runtime"
)

func apiErr(status, code int, message string) *modelruntime.APIError {
	return &modelruntime.APIError{HTTPStatus: status, Code: code, Message: message}
}

type Services struct {
	Auth      *AuthLogic
	User      *UserLogic
	Group     *GroupLogic
	Subject   *SubjectLogic
	SMSConfig *SMSConfigLogic
	AuditLog  *AuditLogLogic
}

func NewServices(core *kernel.Core) *Services {
	return &Services{
		Auth:      &AuthLogic{core: core},
		User:      &UserLogic{core: core},
		Group:     &GroupLogic{core: core},
		Subject:   &SubjectLogic{core: core},
		SMSConfig: &SMSConfigLogic{core: core},
		AuditLog:  &AuditLogLogic{core: core},
	}
}

var _ = http.StatusOK
