package orderlogic

import (
	"net/http"
	"time"

	"myjob/internal/app"
	"myjob/internal/service"
)

const (
	OrderStatusPendingSubmit = "pending_submit"
	OrderStatusProcessing    = "processing"
	OrderStatusSuccess       = "success"
	OrderStatusFailed        = "failed"
	OrderStatusUnknown       = "unknown"

	OrderAttemptStatusPending    = "pending"
	OrderAttemptStatusSubmitted  = "submitted"
	OrderAttemptStatusProcessing = "processing"
	OrderAttemptStatusSuccess    = "success"
	OrderAttemptStatusFailed     = "failed"
	OrderAttemptStatusUnknown    = "unknown"

	ProviderCodeKakayun = "kakayun"
)

// OrderLogic 编排开放订单、上游提交、查单轮询和后台列表。
type OrderLogic struct {
	core             *app.Core
	httpClient       *http.Client
	worker           *Worker
	orderNoGenerator func() string
}

// NewOrderLogic 创建订单业务逻辑实现。
func NewOrderLogic(core *app.Core) *OrderLogic {
	logic := &OrderLogic{
		core:       core,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	logic.orderNoGenerator = logic.generateOrderNo
	return logic
}

var _ service.OrderService = (*OrderLogic)(nil)

// TriggerSubmit 通知订单 worker 尽快扫描待提交订单。
func (l *OrderLogic) TriggerSubmit(orderNo string) {
	if l.worker == nil {
		return
	}
	l.worker.Trigger(orderNo)
}
