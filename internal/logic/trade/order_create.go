package tradelogic

import (
	"context"
	"math/rand"
	"strings"

	"myjob/internal/consts"
	"myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateOrder 创建交易订单，并在事务提交后执行首个 attempt 的真实建单请求。
func (l *TradeOrderLogic) CreateOrder(ctx context.Context, input CreateTradeOrderInput) (entity.TradeOrder, error) {
	input.ClientOrderNo = strings.TrimSpace(input.ClientOrderNo)
	input.GoodsCode = strings.TrimSpace(input.GoodsCode)
	input.PayloadJSON = strings.TrimSpace(input.PayloadJSON)
	if input.CallerID <= 0 {
		return entity.TradeOrder{}, apiErr(consts.CodeBadRequest, "caller_id错误")
	}
	if input.ClientOrderNo == "" {
		return entity.TradeOrder{}, apiErr(consts.CodeBadRequest, "client_order_no不能为空")
	}
	if input.GoodsCode == "" {
		return entity.TradeOrder{}, apiErr(consts.CodeBadRequest, "goods_code不能为空")
	}
	if input.Quantity <= 0 {
		return entity.TradeOrder{}, apiErr(consts.CodeBadRequest, "quantity错误")
	}

	now := l.core.Now()
	if input.RequestedAt.IsZero() {
		input.RequestedAt = now
	}

	var (
		order     entity.TradeOrder
		created   bool
		attemptID int64
	)
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		txRunner := tx.Ctx(ctx)
		existing, ok, err := l.findOrderByClientOrderNo(ctx, txRunner, input.CallerID, input.ClientOrderNo)
		if err != nil {
			return err
		}
		if ok {
			order = existing
			return nil
		}

		prepared, err := buildCandidateBindings(ctx, txRunner, now, input.GoodsCode, input.Quantity, input.PayloadJSON)
		if err != nil {
			return err
		}

		routeRand := rand.New(rand.NewSource(now.UnixNano()))
		firstBinding, err := PickFirstBinding(prepared.Config.RouteMode, prepared.Candidates, now, routeRand)
		if err != nil {
			return err
		}

		orderProvider, ok := l.lookupOrderProvider(firstBinding.ProviderCode)
		if !ok {
			return apiErr(consts.CodeBadRequest, "provider不支持下单")
		}

		plan, err := PlanFulfillments(input.Quantity, orderProvider.SupportsNativeQuantity())
		if err != nil {
			return apiErr(consts.CodeBadRequest, err.Error())
		}

		lockedSalePrice, err := LockSalePrice(prepared.Goods.DefaultSellPrice, firstBinding)
		if err != nil {
			return apiErr(consts.CodeBadRequest, err.Error())
		}
		totalAmount := Round4(lockedSalePrice.Mul(decimal.NewFromInt(int64(input.Quantity))))

		lossAmount, err := EnsureLossAllowed(prepared.Config.AllowLoss, prepared.Config.MaxLossAmount, firstBinding.CostPrice, lockedSalePrice)
		if err != nil {
			return apiErr(consts.CodeBadRequest, "亏本保护拦截")
		}

		orderNo := NewOrderNo(now, now.UnixNano())
		orderID, err := l.insertTradeOrder(ctx, tx, orderNo, input, prepared, firstBinding, lockedSalePrice, totalAmount, lossAmount)
		if err != nil {
			return err
		}

		// 先插入全部 fulfillment 的首个 attempt（A01），但一期仅立即执行第一个 fulfillment。
		for _, item := range plan {
			providerRequestOrderNo := NewProviderRequestOrderNo(orderNo, item.FulfillmentNo, 1)
			attemptRowID, err := l.insertTradeOrderAttempt(ctx, tx, orderID, providerRequestOrderNo, item, firstBinding, prepared.Goods.SubjectName, lockedSalePrice, lossAmount, 1)
			if err != nil {
				return err
			}
			if item.FulfillmentNo == FulfillmentNo(1) {
				attemptID = attemptRowID
			}
		}

		created = true
		order = entity.TradeOrder{
			ID:                orderID,
			OrderNo:           orderNo,
			CallerID:          input.CallerID,
			ClientOrderNo:     input.ClientOrderNo,
			GoodsID:           prepared.Goods.ID,
			GoodsCodeSnapshot: prepared.Goods.GoodsCode,
			GoodsNameSnapshot: prepared.Goods.GoodsName,
			BindingID:         firstBinding.ID,
			PlatformAccountID: firstBinding.PlatformAccountID,
			RouteModeSnapshot: prepared.Config.RouteMode,
			Quantity:          input.Quantity,
			PayloadJSON:       input.PayloadJSON,
			SalePrice:         MoneyString(lockedSalePrice),
			TotalAmount:       MoneyString(totalAmount),
			Status:            "processing",
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		return nil
	}); err != nil {
		return entity.TradeOrder{}, err
	}

	// 幂等命中：直接返回历史订单，不重新选路/锁价/建单。
	if !created {
		return order, nil
	}

	if attemptID > 0 {
		if err := l.executeCreateAttempt(ctx, attemptID, uuid.NewString()); err != nil {
			// 建单失败不回滚订单写库；返回 processing，让异步查单/回调补偿（后续周期补齐）。
			return order, nil
		}
	}
	return order, nil
}

var _ = supplierprovider.CreateOrderInput{}
