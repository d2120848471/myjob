package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_AdminOrderPathExposed(t *testing.T) {
	h := newTestHarness(t)
	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/orders")
}

func TestOrderManagePermissionSeeded(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	me := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, me.Code)
	var data struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(me.Data, &data))
	require.Contains(t, data.Permissions, "order.manage")
}

func TestAdminOrderListRequiresPermission(t *testing.T) {
	h := newTestHarness(t)
	limitedToken := h.createLimitedUserToken(t, h.loginAdmin(t), 0)
	res := h.getJSON("/api/admin/orders?page=1&page_size=20", limitedToken)
	require.Equal(t, 403, res.Code)
}

func TestAdminOrderListFiltersAndStats(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	orderNo := h.insertExternalOrderFixture(t, "O202604240001", "G202604240001", "腾讯视频月卡", "13800138000", "processing", 1)

	res := h.getJSON("/api/admin/orders?page=1&page_size=20&keyword="+orderNo+"&keyword_by=order_no&status=processing&has_tax=1", token)
	require.Equal(t, 0, res.Code)
	var data struct {
		List []struct {
			OrderNo            string `json:"order_no"`
			GoodsID            string `json:"goods_id"`
			SalesSubjectName   string `json:"sales_subject_name"`
			Account            string `json:"account"`
			StatusCode         string `json:"status_code"`
			CurrentChannelName string `json:"current_channel_name"`
			SupplierOrderNo    string `json:"supplier_order_no"`
			AttemptCount       int    `json:"attempt_count"`
		} `json:"list"`
		Stats struct {
			TodayOrderCount  int    `json:"today_order_count"`
			TodayOrderAmount string `json:"today_order_amount"`
		} `json:"stats"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.Len(t, data.List, 1)
	require.Equal(t, orderNo, data.List[0].OrderNo)
	require.Equal(t, "G202604240001", data.List[0].GoodsID)
	require.Equal(t, "测试主体", data.List[0].SalesSubjectName)
	require.Equal(t, "13800138000", data.List[0].Account)
	require.Equal(t, "processing", data.List[0].StatusCode)
	require.Equal(t, "测试云发卡", data.List[0].CurrentChannelName)
	require.Equal(t, "SD202604240001", data.List[0].SupplierOrderNo)
	require.Equal(t, 1, data.List[0].AttemptCount)
	require.GreaterOrEqual(t, data.Stats.TodayOrderCount, 1)
	require.Equal(t, "20.0000", data.Stats.TodayOrderAmount)
}

func (h *testHarness) insertExternalOrderFixture(t *testing.T, orderNo, goodsCode, goodsName, account, status string, hasTax int) string {
	t.Helper()
	now := h.app.Core().Now()
	result, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO external_order (
    order_no, goods_id, goods_code, goods_name, goods_type, supply_type, subject_id, subject_name,
    has_tax, account, quantity, unit_price, order_amount, cost_amount, profit_amount,
    status, attempt_count, last_receipt, created_at, updated_at
) VALUES (?, 1001, ?, ?, 'direct_recharge', 'channel', NULL, '测试主体', ?, ?, 1, '20.0000', '20.0000', '10.0000', '10.0000', ?, 1, '处理中', ?, ?)
`, orderNo, goodsCode, goodsName, hasTax, account, status, now, now)
	require.NoError(t, err)
	orderID, err := result.LastInsertId()
	require.NoError(t, err)

	attemptResult, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO external_order_attempt (
    order_id, order_no, attempt_no, channel_binding_id, platform_account_id, platform_account_name,
    provider_code, supplier_goods_no, supplier_goods_name, supplier_us_order_no, supplier_order_no,
    supplier_status, refund_status, request_snapshot, response_snapshot, receipt, status,
    submitted_at, created_at, updated_at
) VALUES (?, ?, 1, 2001, 3001, '测试云发卡', 'kakayun', '2478510', '腾讯视频月卡', ?, 'SD202604240001', '3', '0', '{}', '{}', '处理中', 'processing', ?, ?, ?)
`, orderID, orderNo, orderNo+"-T1", now, now, now)
	require.NoError(t, err)
	attemptID, err := attemptResult.LastInsertId()
	require.NoError(t, err)
	_, err = h.app.Core().DB().Exec(context.Background(), `UPDATE external_order SET current_attempt_id = ?, updated_at = ? WHERE id = ?`, attemptID, now, orderID)
	require.NoError(t, err)
	return orderNo
}
