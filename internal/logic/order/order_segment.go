package orderlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"myjob/internal/library/channelpricing"
	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

type orderSegmentPlan struct {
	SegmentNo         int
	Quantity          int
	SupplierUSOrderNo string
}

type orderSegmentSubmitPlan struct {
	orderSegmentPlan
	SafetyPrice segmentSafetyPrice
}

type segmentAggregateResult struct {
	OrderStatus     string
	AttemptStatus   string
	SupplierOrderNo string
	Receipt         string
	Message         string
}

func buildOrderSegments(baseSupplierUSOrderNo string, quantity int, capabilities supplierprovider.OrderProviderCapabilities) []orderSegmentPlan {
	if quantity <= 0 {
		return nil
	}
	maxQty := capabilities.MaxQuantityPerCreate
	if maxQty <= 0 || maxQty >= quantity {
		return []orderSegmentPlan{{SegmentNo: 1, Quantity: quantity, SupplierUSOrderNo: baseSupplierUSOrderNo + "-S1"}}
	}
	segments := make([]orderSegmentPlan, 0, (quantity+maxQty-1)/maxQty)
	remaining := quantity
	for remaining > 0 {
		segmentQty := maxQty
		if remaining < segmentQty {
			segmentQty = remaining
		}
		segmentNo := len(segments) + 1
		segments = append(segments, orderSegmentPlan{SegmentNo: segmentNo, Quantity: segmentQty, SupplierUSOrderNo: baseSupplierUSOrderNo + "-S" + intToString(segmentNo)})
		remaining -= segmentQty
	}
	return segments
}

func aggregateSegmentStatuses(segments []entity.ExternalOrderAttemptSegment) segmentAggregateResult {
	if len(segments) == 0 {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusUnknown, AttemptStatus: OrderAttemptStatusUnknown, Message: "上游子单为空"}
	}
	hasUnknown := false
	hasProcessing := false
	hasSubmitted := false
	hasSuccess := false
	hasFailed := false
	hasPending := false
	receipts := make([]string, 0, len(segments))
	supplierOrderNo := ""
	for _, segment := range segments {
		if supplierOrderNo == "" && strings.TrimSpace(segment.SupplierOrderNo) != "" {
			supplierOrderNo = strings.TrimSpace(segment.SupplierOrderNo)
		}
		if strings.TrimSpace(segment.Receipt) != "" {
			receipts = append(receipts, strings.TrimSpace(segment.Receipt))
		}
		switch segment.Status {
		case OrderAttemptStatusFailed:
			hasFailed = true
		case OrderAttemptStatusUnknown:
			hasUnknown = true
		case OrderAttemptStatusProcessing:
			hasProcessing = true
		case OrderAttemptStatusSubmitted:
			hasSubmitted = true
		case OrderAttemptStatusPending, "":
			hasPending = true
		case OrderAttemptStatusSuccess:
			hasSuccess = true
		default:
			hasUnknown = true
		}
	}
	receipt := strings.Join(receipts, "；")
	if hasFailed {
		// 拆单里只要已有子单被受理或状态不确定，就不能把父 attempt 判 failed 去触发整单补单。
		if hasUnknown || hasProcessing || hasSubmitted || hasSuccess {
			message := "上游子单部分失败，需人工确认"
			if receipt != "" {
				message += "：" + receipt
			}
			return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusUnknown, AttemptStatus: OrderAttemptStatusUnknown, SupplierOrderNo: supplierOrderNo, Receipt: message, Message: message}
		}
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusFailed, AttemptStatus: OrderAttemptStatusFailed, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "上游子单失败")}
	}
	if hasUnknown {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusUnknown, AttemptStatus: OrderAttemptStatusUnknown, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "上游子单状态无法确认")}
	}
	if hasProcessing {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusProcessing, AttemptStatus: OrderAttemptStatusProcessing, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "上游子单处理中")}
	}
	if hasSubmitted {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusProcessing, AttemptStatus: OrderAttemptStatusSubmitted, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "上游子单已提交")}
	}
	if hasPending {
		return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusProcessing, AttemptStatus: OrderAttemptStatusPending, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "上游子单等待提交")}
	}
	return segmentAggregateResult{OrderStatus: supplierprovider.SupplierOrderStatusSuccess, AttemptStatus: OrderAttemptStatusSuccess, SupplierOrderNo: supplierOrderNo, Receipt: receipt, Message: defaultOrderMessage(receipt, "全部上游子单成功")}
}

func buildOrderSegmentSubmitPlans(candidate orderChannelCandidate, config reorderConfig, snapshot channelpricing.OrderPriceSnapshot, capabilities supplierprovider.OrderProviderCapabilities, orderQuantity int, segments []orderSegmentPlan) ([]orderSegmentSubmitPlan, error) {
	plans := make([]orderSegmentSubmitPlan, 0, len(segments))
	for _, segment := range segments {
		safetyPrice, err := computeSegmentSafetyPrice(candidate, config, snapshot, capabilities, orderQuantity, segment.Quantity)
		if err != nil {
			return nil, err
		}
		plans = append(plans, orderSegmentSubmitPlan{orderSegmentPlan: segment, SafetyPrice: safetyPrice})
	}
	return plans, nil
}

func insertAttemptSegments(ctx context.Context, tx gdb.TX, attempt entity.ExternalOrderAttempt, candidate orderChannelCandidate, plans []orderSegmentPlan, now time.Time) error {
	for _, plan := range plans {
		if _, err := tx.Exec(`
INSERT INTO external_order_attempt_segment (
    order_id, attempt_id, segment_no, quantity, provider_code, supplier_goods_no,
    supplier_us_order_no, supplier_order_no, supplier_status, refund_status,
    request_snapshot, response_snapshot, receipt, status, submitted_at, last_checked_at,
    created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, '', '', '', '', '', '等待上游提交结果', ?, NULL, NULL, ?, ?)
`, attempt.OrderID, attempt.ID, plan.SegmentNo, plan.Quantity, candidate.ProviderCode, candidate.SupplierGoodsNo,
			plan.SupplierUSOrderNo, OrderAttemptStatusPending, now, now); err != nil {
			return err
		}
	}
	return nil
}

func (l *OrderLogic) submitAttemptSegments(ctx context.Context, provider supplierprovider.OrderProvider, account entity.SupplierPlatformAccount, order entity.ExternalOrder, candidate orderChannelCandidate, attempt entity.ExternalOrderAttempt, baseSupplierUSOrderNo string, plans []orderSegmentSubmitPlan) (createSubmitResult, error) {
	rows, err := l.loadAttemptSegments(ctx, attempt.ID)
	if err != nil {
		return createSubmitResult{}, err
	}
	segmentsByNo := make(map[int]entity.ExternalOrderAttemptSegment, len(rows))
	for _, row := range rows {
		segmentsByNo[row.SegmentNo] = row
	}
	for _, plan := range plans {
		segment, ok := segmentsByNo[plan.SegmentNo]
		if !ok {
			return createSubmitResult{}, fmt.Errorf("上游子单%d不存在", plan.SegmentNo)
		}
		safetyValue := ""
		if plan.SafetyPrice.SendToSupplier {
			safetyValue = plan.SafetyPrice.Value
		}
		result, err := l.executeCreateOrder(ctx, provider, account, supplierprovider.CreateOrderInput{
			SupplierGoodsNo:   candidate.SupplierGoodsNo,
			Quantity:          plan.Quantity,
			Account:           order.Account,
			SupplierUSOrderNo: plan.SupplierUSOrderNo,
			MaxMoney:          safetyValue,
			SafePrice:         safetyValue,
		})
		if err != nil && result.OrderStatus == "" {
			result = createSubmitResult{
				OrderStatus:       OrderStatusUnknown,
				AttemptStatus:     OrderAttemptStatusUnknown,
				SupplierUSOrderNo: plan.SupplierUSOrderNo,
				Message:           err.Error(),
			}
		}
		if strings.TrimSpace(result.SupplierUSOrderNo) == "" {
			result.SupplierUSOrderNo = plan.SupplierUSOrderNo
		}
		if strings.TrimSpace(result.AttemptStatus) == "" {
			result.AttemptStatus = OrderAttemptStatusUnknown
		}
		if err := l.updateSegmentFromCreate(ctx, segment.ID, result); err != nil {
			return createSubmitResult{}, err
		}
		if result.OrderStatus == OrderStatusFailed || result.AttemptStatus == OrderAttemptStatusFailed {
			break
		}
	}
	updated, err := l.loadAttemptSegments(ctx, attempt.ID)
	if err != nil {
		return createSubmitResult{}, err
	}
	aggregate := aggregateSegmentStatuses(updated)
	message := defaultOrderMessage(aggregate.Receipt, aggregate.Message)
	return createSubmitResult{
		OrderStatus:       aggregate.OrderStatus,
		AttemptStatus:     aggregate.AttemptStatus,
		SupplierOrderNo:   aggregate.SupplierOrderNo,
		SupplierUSOrderNo: baseSupplierUSOrderNo,
		Receipt:           message,
		Message:           message,
		RequestSnapshot:   "详见 external_order_attempt_segment",
		ResponseSnapshot:  "详见 external_order_attempt_segment",
	}, nil
}

func (l *OrderLogic) loadAttemptSegments(ctx context.Context, attemptID int64) ([]entity.ExternalOrderAttemptSegment, error) {
	rows := make([]entity.ExternalOrderAttemptSegment, 0)
	err := l.core.DB().GetCore().GetScan(ctx, &rows, `SELECT * FROM external_order_attempt_segment WHERE attempt_id = ? ORDER BY segment_no ASC`, attemptID)
	return rows, err
}

func (l *OrderLogic) pollAttemptSegments(ctx context.Context, order entity.ExternalOrder, attempt entity.ExternalOrderAttempt, provider supplierprovider.OrderProvider, account entity.SupplierPlatformAccount, segments []entity.ExternalOrderAttemptSegment) error {
	for _, segment := range segments {
		if isTerminalSegmentStatus(segment.Status) {
			continue
		}
		result, err := l.executeQueryOrder(provider, account, supplierprovider.QueryOrderInput{
			SupplierOrderNo:   segment.SupplierOrderNo,
			SupplierUSOrderNo: segment.SupplierUSOrderNo,
		})
		if err != nil && result.Status == "" {
			return err
		}
		if strings.TrimSpace(result.SupplierUSOrderNo) == "" {
			result.SupplierUSOrderNo = segment.SupplierUSOrderNo
		}
		if strings.TrimSpace(result.SupplierOrderNo) == "" {
			result.SupplierOrderNo = segment.SupplierOrderNo
		}
		if strings.TrimSpace(result.AttemptStatus) == "" {
			result.AttemptStatus = OrderAttemptStatusUnknown
		}
		if err := l.updateSegmentFromPoll(ctx, segment.ID, result); err != nil {
			return err
		}
	}
	updated, err := l.loadAttemptSegments(ctx, attempt.ID)
	if err != nil {
		return err
	}
	aggregate := aggregateSegmentStatuses(updated)
	message := defaultOrderMessage(aggregate.Receipt, aggregate.Message)
	result := queryPollResult{
		Status:            aggregate.OrderStatus,
		AttemptStatus:     aggregate.AttemptStatus,
		SupplierOrderNo:   aggregate.SupplierOrderNo,
		SupplierUSOrderNo: attempt.SupplierUSOrderNo,
		Receipt:           message,
		Message:           message,
		ResponseSnapshot:  "详见 external_order_attempt_segment",
	}
	switch result.Status {
	case supplierprovider.SupplierOrderStatusSuccess:
		return l.applyPollSuccess(ctx, order, attempt, result)
	case supplierprovider.SupplierOrderStatusProcessing:
		return l.applyPollProcessing(ctx, order, attempt, result)
	case supplierprovider.SupplierOrderStatusFailed:
		if err := l.applyPollAttemptFailed(ctx, order, attempt, result); err != nil {
			return err
		}
		return l.handleAttemptFailed(ctx, order, attempt, defaultOrderMessage(result.Receipt, result.Message))
	default:
		return l.applyPollUnknown(ctx, order, attempt, result)
	}
}

func (l *OrderLogic) updateSegmentFromCreate(ctx context.Context, segmentID int64, result createSubmitResult) error {
	now := l.core.Now()
	_, err := l.core.DB().Exec(ctx, `
UPDATE external_order_attempt_segment
SET supplier_order_no = ?, supplier_us_order_no = ?, supplier_status = ?, refund_status = ?,
    request_snapshot = ?, response_snapshot = ?, receipt = ?, status = ?, submitted_at = ?, updated_at = ?
WHERE id = ?
`, result.SupplierOrderNo, result.SupplierUSOrderNo, result.SupplierStatus, result.RefundStatus,
		result.RequestSnapshot, result.ResponseSnapshot, defaultOrderMessage(result.Receipt, result.Message), result.AttemptStatus, now, now, segmentID)
	return err
}

func (l *OrderLogic) updateSegmentFromPoll(ctx context.Context, segmentID int64, result queryPollResult) error {
	now := l.core.Now()
	_, err := l.core.DB().Exec(ctx, `
UPDATE external_order_attempt_segment
SET supplier_order_no = ?, supplier_us_order_no = ?, supplier_status = ?, refund_status = ?,
    response_snapshot = ?, receipt = ?, status = ?, last_checked_at = ?, updated_at = ?
WHERE id = ?
`, result.SupplierOrderNo, result.SupplierUSOrderNo, result.SupplierStatus, result.RefundStatus,
		result.ResponseSnapshot, defaultOrderMessage(result.Receipt, result.Message), result.AttemptStatus, now, now, segmentID)
	return err
}

func isTerminalSegmentStatus(status string) bool {
	return status == OrderAttemptStatusSuccess || status == OrderAttemptStatusFailed
}
