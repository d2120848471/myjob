package tradelogic

import (
	"context"
	"net/http"
	"time"

	"myjob/internal/app"
	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"
)

// TradeOrderLogic 提供交易订单创建/查单等核心能力（不包含管理端绑定维护）。
type TradeOrderLogic struct {
	core   *app.Core
	client *http.Client

	lookupOrderProvider func(code string) (supplierprovider.OrderProvider, bool)
}

// NewTradeOrderLogic 创建 TradeOrderLogic。
func NewTradeOrderLogic(core *app.Core, lookup func(code string) (supplierprovider.OrderProvider, bool), client *http.Client) *TradeOrderLogic {
	if lookup == nil {
		lookup = supplierprovider.LookupOrder
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &TradeOrderLogic{core: core, lookupOrderProvider: lookup, client: client}
}

func (l *TradeOrderLogic) findOrderByClientOrderNo(ctx context.Context, runner sqlRunner, callerID int64, clientOrderNo string) (entity.TradeOrder, bool, error) {
	record, err := runner.GetOne(`
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
    created_at,
    updated_at
FROM trade_order
WHERE caller_id = ? AND client_order_no = ?
LIMIT 1
`, callerID, clientOrderNo)
	if err != nil {
		return entity.TradeOrder{}, false, apiErr(consts.CodeInternalError, "读取订单失败")
	}
	if record == nil || len(record) == 0 {
		return entity.TradeOrder{}, false, nil
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
		CreatedAt:         record["created_at"].Time(),
		UpdatedAt:         record["updated_at"].Time(),
	}, true, nil
}
