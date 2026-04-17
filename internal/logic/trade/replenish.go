package tradelogic

import (
	"context"
	"math/rand"
	"strings"

	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
)

type attemptBindingIDRow struct {
	BindingID int64 `db:"binding_id"`
}

func loadAttemptedBindingIDs(ctx context.Context, runner sqlRunner, orderID int64, fulfillmentNo string) (map[int64]struct{}, error) {
	fulfillmentNo = strings.TrimSpace(fulfillmentNo)
	if orderID <= 0 || fulfillmentNo == "" {
		return map[int64]struct{}{}, nil
	}
	rows := make([]attemptBindingIDRow, 0)
	if err := runner.GetScan(&rows, `
SELECT DISTINCT binding_id
FROM trade_order_attempt
WHERE order_id = ? AND fulfillment_no = ?
`, orderID, fulfillmentNo); err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取attempt失败")
	}
	result := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row.BindingID > 0 {
			result[row.BindingID] = struct{}{}
		}
	}
	return result, nil
}

// tryReplenishAfterFailedAttempt 在 attempt 明确失败时，尝试按候选绑定切换到下一条绑定继续建单。
func (l *TradeOrderLogic) tryReplenishAfterFailedAttempt(ctx context.Context, current createAttemptRow, traceID string) (int64, bool, error) {
	if current.OrderID <= 0 || strings.TrimSpace(current.GoodsCode) == "" || strings.TrimSpace(current.FulfillmentNo) == "" {
		return 0, false, nil
	}
	if current.OrderQuantity <= 0 || current.AttemptQuantity <= 0 {
		return 0, false, nil
	}

	now := l.core.Now()
	nextAttemptID := int64(0)
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		txRunner := tx.Ctx(ctx)

		prepared, err := buildCandidateBindings(ctx, txRunner, now, current.GoodsCode, current.OrderQuantity, current.PayloadJSON)
		if err != nil {
			return err
		}
		if !prepared.Config.SmartReplenishEnabled {
			return nil
		}

		attemptedIDs, err := loadAttemptedBindingIDs(ctx, txRunner, current.OrderID, current.FulfillmentNo)
		if err != nil {
			return err
		}

		lockedSalePrice, err := ParseMoney(current.SalePrice)
		if err != nil {
			return apiErr(consts.CodeInternalError, "订单售价快照错误")
		}

		routeRand := rand.New(rand.NewSource(now.UnixNano()))
		for i := 0; i < len(prepared.Candidates)+2; i++ {
			nextBinding, ok, pickErr := PickNextBinding(prepared.Config.RouteMode, prepared.Candidates, now, attemptedIDs, routeRand)
			if pickErr != nil {
				return pickErr
			}
			if !ok {
				return nil
			}

			if _, ok := l.lookupOrderProvider(strings.TrimSpace(nextBinding.ProviderCode)); !ok {
				attemptedIDs[nextBinding.ID] = struct{}{}
				continue
			}

			lossAmount, lossErr := EnsureLossAllowed(prepared.Config.AllowLoss, prepared.Config.MaxLossAmount, nextBinding.CostPrice, lockedSalePrice)
			if lossErr != nil {
				attemptedIDs[nextBinding.ID] = struct{}{}
				continue
			}

			maxAttemptNo, err := txRunner.GetValue(`
SELECT COALESCE(MAX(attempt_no), 0)
FROM trade_order_attempt
WHERE order_id = ? AND fulfillment_no = ?
`, current.OrderID, current.FulfillmentNo)
			if err != nil {
				return apiErr(consts.CodeInternalError, "读取attempt失败")
			}
			attemptNo := maxAttemptNo.Int() + 1
			if attemptNo <= 1 {
				attemptNo = 2
			}

			providerRequestOrderNo := NewProviderRequestOrderNo(strings.TrimSpace(current.OrderNo), strings.TrimSpace(current.FulfillmentNo), attemptNo)
			plan := FulfillmentPlanItem{FulfillmentNo: strings.TrimSpace(current.FulfillmentNo), AttemptQuantity: current.AttemptQuantity}

			newID, err := l.insertTradeOrderAttempt(ctx, tx, current.OrderID, providerRequestOrderNo, plan, nextBinding, prepared.Goods.SubjectName, lockedSalePrice, lossAmount, attemptNo)
			if err != nil {
				return err
			}
			nextAttemptID = newID
			return nil
		}

		return nil
	}); err != nil {
		return 0, false, err
	}

	if nextAttemptID <= 0 {
		return 0, false, nil
	}
	return nextAttemptID, true, nil
}
