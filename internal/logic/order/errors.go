package orderlogic

import (
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

func apiErr(code gcode.Code, message string) error {
	return gerror.NewCode(code, message)
}

func unauthorizedErr() error {
	return apiErr(consts.CodeUnauthorized, "token错误")
}
