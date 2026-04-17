package service

import (
	"context"
	"net/http"
)

// TradePriceNotifyService 定义上游价格通知处理能力。
type TradePriceNotifyService interface {
	HandleProviderPriceNotify(ctx context.Context, providerCode string, headers http.Header, body []byte) ([]byte, string, error)
}

