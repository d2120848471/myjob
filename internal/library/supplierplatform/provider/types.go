package supplierprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// PlatformType 描述第三方平台类型字典项（用于初始化 supplier_platform_type 表）。
type PlatformType struct {
	ID           int
	TypeName     string
	ProviderCode string
	Status       int
	Sort         int
}

// AccountConfig 描述第三方平台账号的请求配置（域名/凭证/扩展配置等）。
type AccountConfig struct {
	ProviderCode string
	Domain       string
	BackupDomain string
	TokenID      string
	SecretKey    string
	ExtraConfig  map[string]any
}

// BalanceProvider 定义第三方平台余额查询适配器能力。
type BalanceProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error)
	ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error)
}
