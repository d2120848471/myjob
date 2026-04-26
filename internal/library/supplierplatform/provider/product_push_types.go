package supplierprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// ProductSubscribeInput 描述订阅或取消订阅单个上游商品所需参数。
type ProductSubscribeInput struct {
	SupplierGoodsNo string
}

// ProductChangePushResult 表示供应商商品变动推送解析后的稳定数据。
type ProductChangePushResult struct {
	SupplierGoodsNo string
	GoodsName       string
	GoodsPrice      decimal.Decimal
	GoodsPriceValid bool
	GoodsStatus     string
	Raw             string
}

// ProductSubscriptionProvider 定义供应商商品推送订阅能力。
type ProductSubscriptionProvider interface {
	Code() string
	Name() string
	BuildSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error)
	BuildCancelSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error)
	ParseMutationResponse(statusCode int, body []byte) (string, error)
}

// ProductChangePushProvider 定义供应商商品变动推送验签和解析能力。
type ProductChangePushProvider interface {
	Code() string
	Name() string
	ParseProductChangePush(account AccountConfig, now time.Time, body []byte) (ProductChangePushResult, error)
}
