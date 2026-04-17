package tradelogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	openapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/shopspring/decimal"
)

// OpenOrderLogic 实现开放接口下单与查单：鉴权由 middleware 完成，本逻辑只做 DTO 装配与交易核心调用。
type OpenOrderLogic struct {
	core  *app.Core
	trade *TradeOrderLogic
}

// NewOpenOrderLogic 创建 OpenOrderLogic。
func NewOpenOrderLogic(core *app.Core) *OpenOrderLogic {
	return &OpenOrderLogic{core: core, trade: NewTradeOrderLogic(core, nil, nil)}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func mapOpenOrderStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "success":
		return "success"
	case "failed", "manual_review":
		return "failed"
	default:
		return "processing"
	}
}

func mapAttemptStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "success":
		return "success"
	case "failed", "timeout":
		return "failed"
	default:
		return "processing"
	}
}

// Create 创建订单（幂等）。
func (l *OpenOrderLogic) Create(ctx context.Context, req *openapi.OpenOrderCreateReq, caller entity.OpenCaller, ip string) (*openapi.OpenOrderCreateRes, error) {
	if caller.ID <= 0 {
		return nil, apiErr(consts.CodeUnauthorized, "caller_id错误")
	}
	payloadJSON := "{}"
	if req.Payload != nil {
		raw, _ := json.Marshal(req.Payload)
		if len(raw) > 0 {
			payloadJSON = string(raw)
		}
	}

	order, err := l.trade.CreateOrder(ctx, CreateTradeOrderInput{
		CallerID:      caller.ID,
		ClientOrderNo: strings.TrimSpace(req.ClientOrderNo),
		GoodsCode:     strings.TrimSpace(req.GoodsCode),
		Quantity:      req.Quantity,
		PayloadJSON:   payloadJSON,
		RequestIP:     strings.TrimSpace(ip),
		RequestedAt:   time.Time{},
	})
	if err != nil {
		return nil, err
	}
	return &openapi.OpenOrderCreateRes{
		OrderNo:       strings.TrimSpace(order.OrderNo),
		ClientOrderNo: strings.TrimSpace(order.ClientOrderNo),
		Status:        mapOpenOrderStatus(order.Status),
		GoodsCode:     strings.TrimSpace(order.GoodsCodeSnapshot),
		GoodsName:     strings.TrimSpace(order.GoodsNameSnapshot),
		Quantity:      order.Quantity,
		SalePrice:     strings.TrimSpace(order.SalePrice),
		TotalAmount:   strings.TrimSpace(order.TotalAmount),
		CreatedAt:     formatTime(order.CreatedAt),
	}, nil
}

// Get 按内部订单号查单。
func (l *OpenOrderLogic) Get(ctx context.Context, req *openapi.OpenOrderGetReq, caller entity.OpenCaller) (*openapi.OpenOrderGetRes, error) {
	order, ok, err := l.loadOrderByOrderNo(ctx, caller.ID, strings.TrimSpace(req.OrderNo))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apiErr(consts.CodeBadRequest, "订单不存在")
	}
	return l.buildOpenOrderRes(ctx, order)
}

// GetByClient 按调用方订单号查单。
func (l *OpenOrderLogic) GetByClient(ctx context.Context, req *openapi.OpenOrderGetByClientReq, caller entity.OpenCaller) (*openapi.OpenOrderGetRes, error) {
	order, ok, err := l.trade.findOrderByClientOrderNo(ctx, ctxDBRunner{ctx: ctx, db: l.core.DB()}, caller.ID, strings.TrimSpace(req.ClientOrderNo))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apiErr(consts.CodeBadRequest, "订单不存在")
	}
	return l.buildOpenOrderRes(ctx, order)
}

func (l *OpenOrderLogic) loadOrderByOrderNo(ctx context.Context, callerID int64, orderNo string) (entity.TradeOrder, bool, error) {
	orderNo = strings.TrimSpace(orderNo)
	if callerID <= 0 || orderNo == "" {
		return entity.TradeOrder{}, false, nil
	}
	record, err := l.core.DB().GetCore().GetOne(ctx, `
SELECT
    id,
    order_no,
    caller_id,
    client_order_no,
    goods_id,
    goods_code_snapshot,
    goods_name_snapshot,
    quantity,
    success_quantity,
    failed_quantity,
    sale_price,
    total_amount,
    status,
    failure_reason,
    finished_at,
    created_at,
    updated_at
FROM trade_order
WHERE caller_id = ? AND order_no = ?
LIMIT 1
`, callerID, orderNo)
	if err != nil {
		return entity.TradeOrder{}, false, apiErr(consts.CodeInternalError, "读取订单失败")
	}
	if record == nil || len(record) == 0 {
		return entity.TradeOrder{}, false, nil
	}

	finishedAt := sql.NullTime{}
	if record["finished_at"] != nil && !record["finished_at"].IsNil() {
		value := record["finished_at"].Time()
		if !value.IsZero() {
			finishedAt = sql.NullTime{Time: value, Valid: true}
		}
	}
	return entity.TradeOrder{
		ID:                record["id"].Int64(),
		OrderNo:           record["order_no"].String(),
		CallerID:          record["caller_id"].Int64(),
		ClientOrderNo:     record["client_order_no"].String(),
		GoodsID:           record["goods_id"].Int64(),
		GoodsCodeSnapshot: record["goods_code_snapshot"].String(),
		GoodsNameSnapshot: record["goods_name_snapshot"].String(),
		Quantity:          record["quantity"].Int(),
		SuccessQuantity:   record["success_quantity"].Int(),
		FailedQuantity:    record["failed_quantity"].Int(),
		SalePrice:         record["sale_price"].String(),
		TotalAmount:       record["total_amount"].String(),
		Status:            record["status"].String(),
		FailureReason:     record["failure_reason"].String(),
		FinishedAt:        finishedAt,
		CreatedAt:         record["created_at"].Time(),
		UpdatedAt:         record["updated_at"].Time(),
	}, true, nil
}

func (l *OpenOrderLogic) buildOpenOrderRes(ctx context.Context, order entity.TradeOrder) (*openapi.OpenOrderGetRes, error) {
	attempts := make([]openapi.OpenUpstreamOrderItem, 0)
	type row struct {
		FulfillmentNo                  string `db:"fulfillment_no"`
		AttemptNo                      int    `db:"attempt_no"`
		AttemptQuantity                int    `db:"attempt_quantity"`
		AttemptStatus                  string `db:"attempt_status"`
		UpstreamStatus                 string `db:"upstream_status"`
		ProviderRequestOrderNo         string `db:"provider_request_order_no"`
		ChannelOrderNo                 string `db:"channel_order_no"`
		BindingChannelNameSnapshot     string `db:"binding_channel_name_snapshot"`
		BindingSupplierGoodsNoSnapshot string `db:"binding_supplier_goods_no_snapshot"`
		ErrorCategory                  string `db:"error_category"`
		ErrorMessage                   string `db:"error_message"`
		FinishedAt                     time.Time `db:"finished_at"`
	}
	rows := make([]row, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    fulfillment_no,
    attempt_no,
    attempt_quantity,
    attempt_status,
    upstream_status,
    provider_request_order_no,
    channel_order_no,
    binding_channel_name_snapshot,
    binding_supplier_goods_no_snapshot,
    error_category,
    error_message,
    finished_at
FROM trade_order_attempt
WHERE order_id = ?
ORDER BY fulfillment_no ASC, attempt_no ASC, id ASC
`, order.ID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取attempt失败")
	}
	for _, item := range rows {
		attempts = append(attempts, openapi.OpenUpstreamOrderItem{
			FulfillmentNo:          strings.TrimSpace(item.FulfillmentNo),
			AttemptNo:              item.AttemptNo,
			AttemptQuantity:        item.AttemptQuantity,
			Status:                 mapAttemptStatus(item.AttemptStatus),
			BindingChannelName:     strings.TrimSpace(item.BindingChannelNameSnapshot),
			BindingSupplierGoodsNo: strings.TrimSpace(item.BindingSupplierGoodsNoSnapshot),
			ChannelOrderNo:         strings.TrimSpace(item.ChannelOrderNo),
			ProviderRequestOrderNo: strings.TrimSpace(item.ProviderRequestOrderNo),
			UpstreamStatus:         strings.TrimSpace(item.UpstreamStatus),
			ErrorCategory:          strings.TrimSpace(item.ErrorCategory),
			ErrorMessage:           strings.TrimSpace(item.ErrorMessage),
			FinishedAt:             formatTime(item.FinishedAt),
		})
	}

	finishedAt := ""
	if order.FinishedAt.Valid {
		finishedAt = formatTime(order.FinishedAt.Time)
	}

	// 所有金额继续以字符串形式透出，保持小数精度一致性。
	_, _ = decimal.NewFromString(order.SalePrice)
	_, _ = decimal.NewFromString(order.TotalAmount)

	return &openapi.OpenOrderGetRes{
		OrderNo:         strings.TrimSpace(order.OrderNo),
		ClientOrderNo:   strings.TrimSpace(order.ClientOrderNo),
		Status:          mapOpenOrderStatus(order.Status),
		GoodsCode:       strings.TrimSpace(order.GoodsCodeSnapshot),
		GoodsName:       strings.TrimSpace(order.GoodsNameSnapshot),
		Quantity:        order.Quantity,
		SuccessQuantity: order.SuccessQuantity,
		FailedQuantity:  order.FailedQuantity,
		SalePrice:       strings.TrimSpace(order.SalePrice),
		TotalAmount:     strings.TrimSpace(order.TotalAmount),
		FailureReason:   strings.TrimSpace(order.FailureReason),
		CreatedAt:       formatTime(order.CreatedAt),
		FinishedAt:      finishedAt,
		UpstreamOrders:  attempts,
	}, nil
}
