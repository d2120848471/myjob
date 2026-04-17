package supplierprovider

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// PriceNotifyResult 是 Provider 解析价格通知后的统一结果对象。
type PriceNotifyResult struct {
	PlatformAccountLocator string
	SupplierGoodsNo        string
	SupplierGoodsName      string
	SourceCostPrice        decimal.Decimal
	NotifyAt               time.Time
	IdempotencyKey          string
	RawPayload              string
}

// PriceNotifyProvider 定义上游价格通知验签与解析能力。
type PriceNotifyProvider interface {
	Code() string
	Name() string
	VerifyPriceNotifySignature(account AccountConfig, headers http.Header, body []byte) error
	ParsePriceNotifyPayload(account AccountConfig, headers http.Header, body []byte) (*PriceNotifyResult, error)
}

