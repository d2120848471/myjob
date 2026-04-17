package service

import (
	"context"
	"net/http"
)

// TradeCallbackService 定义上游订单回调处理能力。
type TradeCallbackService interface {
	HandleProviderOrderCallback(ctx context.Context, providerCode string, headers http.Header, body []byte) ([]byte, string, error)
}

