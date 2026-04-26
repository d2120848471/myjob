package supplierprovider

import (
	"context"
	"errors"
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

var ErrSupplierUnknownResponse = errors.New("supplier response is not confirmed")

const (
	SupplierOrderStatusProcessing = "processing"
	SupplierOrderStatusSuccess    = "success"
	SupplierOrderStatusFailed     = "failed"
	SupplierOrderStatusUnknown    = "unknown"
)

// CreateOrderInput 是上游下单接口所需的最小业务参数。
type CreateOrderInput struct {
	SupplierGoodsNo   string
	Quantity          int
	Account           string
	SupplierUSOrderNo string
	MaxMoney          string
}

// CreateOrderResult 表示上游下单响应解析后的稳定结果。
type CreateOrderResult struct {
	Accepted          bool
	Status            string
	SupplierOrderNo   string
	SupplierUSOrderNo string
	SupplierStatus    string
	Message           string
	Raw               string
}

// QueryOrderInput 是上游查单接口所需的订单定位参数。
type QueryOrderInput struct {
	SupplierOrderNo   string
	SupplierUSOrderNo string
}

// QueryOrderResult 表示上游查单响应解析后的稳定结果。
type QueryOrderResult struct {
	Status            string
	SupplierOrderNo   string
	SupplierUSOrderNo string
	SupplierStatus    string
	RefundStatus      string
	Receipt           string
	Message           string
	Raw               string
}

// OrderProvider 定义第三方平台下单和查单适配器能力。
type OrderProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error)
	ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error)
	BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error)
	ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error)
}

// ProductInfoInput 是上游商品详情接口所需的商品定位参数。
type ProductInfoInput struct {
	SupplierGoodsNo string
}

// ProductInfoResult 表示上游商品详情响应解析后的稳定商品信息。
type ProductInfoResult struct {
	SupplierGoodsNo string
	GoodsName       string
	GoodsPrice      decimal.Decimal
	GoodsPriceValid bool
	Raw             string
}

// ProductInfoProvider 定义第三方平台商品详情查询适配器能力。
type ProductInfoProvider interface {
	Code() string
	Name() string
	CandidateBaseURLs(account AccountConfig) []string
	BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error)
	ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error)
}
