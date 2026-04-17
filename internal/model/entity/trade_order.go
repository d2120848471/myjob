package entity

import (
	"database/sql"
	"time"
)

// TradeOrder 对应 trade_order 主订单表。
type TradeOrder struct {
	ID                      int64        `db:"id" json:"id"`
	OrderNo                 string       `db:"order_no" json:"order_no"`
	CallerID                int64        `db:"caller_id" json:"caller_id"`
	ClientOrderNo           string       `db:"client_order_no" json:"client_order_no"`
	GoodsID                 int64        `db:"goods_id" json:"goods_id"`
	GoodsCodeSnapshot       string       `db:"goods_code_snapshot" json:"goods_code_snapshot"`
	GoodsNameSnapshot       string       `db:"goods_name_snapshot" json:"goods_name_snapshot"`
	BindingID               int64        `db:"binding_id" json:"binding_id"`
	PlatformAccountID       int64        `db:"platform_account_id" json:"platform_account_id"`
	RouteModeSnapshot       string       `db:"route_mode_snapshot" json:"route_mode_snapshot"`
	Quantity                int          `db:"quantity" json:"quantity"`
	SuccessQuantity         int          `db:"success_quantity" json:"success_quantity"`
	FailedQuantity          int          `db:"failed_quantity" json:"failed_quantity"`
	PayloadJSON             string       `db:"payload_json" json:"payload_json"`
	SalePrice               string       `db:"sale_price" json:"sale_price"`
	TotalAmount             string       `db:"total_amount" json:"total_amount"`
	SourceCostPriceSnapshot string       `db:"source_cost_price_snapshot" json:"source_cost_price_snapshot"`
	CostPriceSnapshot       string       `db:"cost_price_snapshot" json:"cost_price_snapshot"`
	TaxAdjustDirection      string       `db:"tax_adjust_direction" json:"tax_adjust_direction"`
	TaxAdjustRate           string       `db:"tax_adjust_rate" json:"tax_adjust_rate"`
	TaxAdjustAmount         string       `db:"tax_adjust_amount" json:"tax_adjust_amount"`
	LossOrder               int          `db:"loss_order" json:"loss_order"`
	LossAmount              string       `db:"loss_amount" json:"loss_amount"`
	ChannelOrderNo          string       `db:"channel_order_no" json:"channel_order_no"`
	Status                  string       `db:"status" json:"status"`
	FailureReason           string       `db:"failure_reason" json:"failure_reason"`
	FinishedAt              sql.NullTime `db:"finished_at" json:"finished_at"`
	CreatedAt               time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time    `db:"updated_at" json:"updated_at"`
}

// TradeOrderAttempt 对应 trade_order_attempt 履约尝试表。
type TradeOrderAttempt struct {
	ID                             int64        `db:"id" json:"id"`
	OrderID                        int64        `db:"order_id" json:"order_id"`
	BindingID                      int64        `db:"binding_id" json:"binding_id"`
	PlatformAccountID              int64        `db:"platform_account_id" json:"platform_account_id"`
	ProviderCode                   string       `db:"provider_code" json:"provider_code"`
	FulfillmentNo                  string       `db:"fulfillment_no" json:"fulfillment_no"`
	AttemptQuantity                int          `db:"attempt_quantity" json:"attempt_quantity"`
	AttemptNo                      int          `db:"attempt_no" json:"attempt_no"`
	ProviderRequestOrderNo         string       `db:"provider_request_order_no" json:"provider_request_order_no"`
	ChannelOrderNo                 string       `db:"channel_order_no" json:"channel_order_no"`
	AttemptStatus                  string       `db:"attempt_status" json:"attempt_status"`
	UpstreamStatus                 string       `db:"upstream_status" json:"upstream_status"`
	BindingChannelNameSnapshot     string       `db:"binding_channel_name_snapshot" json:"binding_channel_name_snapshot"`
	BindingSupplierGoodsNoSnapshot string       `db:"binding_supplier_goods_no_snapshot" json:"binding_supplier_goods_no_snapshot"`
	SourceCostPriceSnapshot        string       `db:"source_cost_price_snapshot" json:"source_cost_price_snapshot"`
	CostPriceSnapshot              string       `db:"cost_price_snapshot" json:"cost_price_snapshot"`
	SalePriceSnapshot              string       `db:"sale_price_snapshot" json:"sale_price_snapshot"`
	LossAmountSnapshot             string       `db:"loss_amount_snapshot" json:"loss_amount_snapshot"`
	RequestURL                     string       `db:"request_url" json:"request_url"`
	RequestMethod                  string       `db:"request_method" json:"request_method"`
	RequestHeaders                 string       `db:"request_headers" json:"request_headers"`
	RequestPayload                 string       `db:"request_payload" json:"request_payload"`
	ResponsePayload                string       `db:"response_payload" json:"response_payload"`
	HTTPStatus                     int          `db:"http_status" json:"http_status"`
	DurationMS                     int          `db:"duration_ms" json:"duration_ms"`
	ErrorCategory                  string       `db:"error_category" json:"error_category"`
	ErrorCode                      string       `db:"error_code" json:"error_code"`
	ErrorMessage                   string       `db:"error_message" json:"error_message"`
	QueryCount                     int          `db:"query_count" json:"query_count"`
	LastQueryAt                    sql.NullTime `db:"last_query_at" json:"last_query_at"`
	NextQueryAt                    sql.NullTime `db:"next_query_at" json:"next_query_at"`
	QueryDeadlineAt                sql.NullTime `db:"query_deadline_at" json:"query_deadline_at"`
	CallbackPayload                string       `db:"callback_payload" json:"callback_payload"`
	CallbackReceivedAt             sql.NullTime `db:"callback_received_at" json:"callback_received_at"`
	CallbackProcessedAt            sql.NullTime `db:"callback_processed_at" json:"callback_processed_at"`
	TraceID                        string       `db:"trace_id" json:"trace_id"`
	FinishedAt                     sql.NullTime `db:"finished_at" json:"finished_at"`
	CreatedAt                      time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt                      time.Time    `db:"updated_at" json:"updated_at"`
}

// ProviderCallbackLog 对应 provider_callback_log 上游回调日志表。
type ProviderCallbackLog struct {
	ID                     int64     `db:"id" json:"id"`
	ProviderCode           string    `db:"provider_code" json:"provider_code"`
	PlatformAccountID      int64     `db:"platform_account_id" json:"platform_account_id"`
	IdempotencyKey         string    `db:"idempotency_key" json:"idempotency_key"`
	ProviderRequestOrderNo string    `db:"provider_request_order_no" json:"provider_request_order_no"`
	ChannelOrderNo         string    `db:"channel_order_no" json:"channel_order_no"`
	RequestHeaders         string    `db:"request_headers" json:"request_headers"`
	RequestBody            string    `db:"request_body" json:"request_body"`
	VerifyResult           string    `db:"verify_result" json:"verify_result"`
	ProcessResult          string    `db:"process_result" json:"process_result"`
	AckBody                string    `db:"ack_body" json:"ack_body"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
}

// ProviderPriceNotifyLog 对应 provider_price_notify_log 上游价格通知日志表。
type ProviderPriceNotifyLog struct {
	ID                 int64     `db:"id" json:"id"`
	ProviderCode       string    `db:"provider_code" json:"provider_code"`
	PlatformAccountID  int64     `db:"platform_account_id" json:"platform_account_id"`
	IdempotencyKey     string    `db:"idempotency_key" json:"idempotency_key"`
	SupplierGoodsNo    string    `db:"supplier_goods_no" json:"supplier_goods_no"`
	RequestHeaders     string    `db:"request_headers" json:"request_headers"`
	RequestBody        string    `db:"request_body" json:"request_body"`
	SourceCostPriceNew string    `db:"source_cost_price_new" json:"source_cost_price_new"`
	VerifyResult       string    `db:"verify_result" json:"verify_result"`
	ProcessResult      string    `db:"process_result" json:"process_result"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
