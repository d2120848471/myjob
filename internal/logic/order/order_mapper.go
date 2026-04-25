package orderlogic

import (
	"strings"
	"time"

	adminapi "myjob/api"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

func orderStatusText(status string) string {
	switch status {
	case OrderStatusPendingSubmit:
		return "待提交"
	case OrderStatusProcessing:
		return "处理中"
	case OrderStatusSuccess:
		return "成功"
	case OrderStatusFailed:
		return "失败"
	case OrderStatusUnknown:
		return "未知"
	default:
		return "未知"
	}
}

func formatAppTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func adminOrderListItemFromRecord(row gdb.Record) adminapi.OrderListItem {
	status := row["status"].String()
	return adminapi.OrderListItem{
		ID:                 row["id"].Int64(),
		SalesSubjectName:   row["sales_subject_name"].String(),
		OrderNo:            row["order_no"].String(),
		GoodsID:            row["goods_code"].String(),
		GoodsName:          row["goods_name"].String(),
		Account:            row["account"].String(),
		Quantity:           row["quantity"].Int(),
		OrderAmount:        formatOrderMoney(row["order_amount"].String()),
		CostAmount:         formatOrderMoney(row["cost_amount"].String()),
		ProfitAmount:       formatOrderMoney(row["profit_amount"].String()),
		CurrentChannelID:   row["current_channel_id"].Int64(),
		CurrentChannelName: row["current_channel_name"].String(),
		SupplierOrderNo:    row["supplier_order_no"].String(),
		AttemptCount:       row["attempt_count"].Int(),
		LastReceipt:        row["last_receipt"].String(),
		StatusCode:         status,
		StatusText:         orderStatusText(status),
		CreatedAt:          formatAppTime(row["created_at"].Time()),
		UpdatedAt:          formatAppTime(row["updated_at"].Time()),
	}
}

func formatOrderMoney(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0.0000"
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return value
	}
	return amount.Round(4).StringFixed(4)
}
