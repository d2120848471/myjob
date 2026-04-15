package adminlogic

import (
	"context"
	"net/http"
	"time"

	"myjob/internal/app"
)

// PlatformDeleteHook 允许在删除第三方平台账号前执行自定义清理逻辑。
type PlatformDeleteHook interface {
	BeforeDelete(ctx context.Context, platformID int64) error
}

type noopPlatformDeleteHook struct{}

func (noopPlatformDeleteHook) BeforeDelete(context.Context, int64) error { return nil }

// SupplierPlatformLogic 提供第三方平台账号管理与余额查询相关业务能力。
type SupplierPlatformLogic struct {
	core       *app.Core
	httpClient *http.Client
	deleteHook PlatformDeleteHook
}

// NewSupplierPlatformLogic 创建 SupplierPlatformLogic，并初始化默认 HTTP client 与删除钩子。
func NewSupplierPlatformLogic(core *app.Core) *SupplierPlatformLogic {
	return &SupplierPlatformLogic{
		core:       core,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		deleteHook: noopPlatformDeleteHook{},
	}
}
