package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_RechargeRiskPathsExposed(t *testing.T) {
	h := newTestHarness(t)
	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/recharge-risks/rules")
	require.Contains(t, res.body, "/api/admin/recharge-risks/records")
}

func TestRechargeRiskPermissionSeeded(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	me := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, me.Code)
	var data struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(me.Data, &data))
	require.Contains(t, data.Permissions, "order.recharge_risk")
}

func TestRechargeRiskRequiresPermission(t *testing.T) {
	h := newTestHarness(t)
	limitedToken := h.createLimitedUserToken(t, h.loginAdmin(t), 0)
	res := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20", limitedToken)
	require.Equal(t, 403, res.Code)
}

func TestRechargeRiskRuleCRUDAndFilters(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	invalid := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":       "",
		"goods_keyword": "剪映",
		"reason":        "测试原因",
		"status":        1,
	}, token)
	require.Equal(t, 400, invalid.Code)

	create := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":       "risk-account-001",
		"goods_keyword": "剪映",
		"reason":        "客户多次提交错误账号",
		"status":        1,
	}, token)
	require.Equal(t, 0, create.Code)
	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotZero(t, createData.ID)

	duplicate := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":       "risk-account-001",
		"goods_keyword": "剪映",
		"reason":        "重复",
		"status":        1,
	}, token)
	require.Equal(t, 409, duplicate.Code)

	list := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&account=risk-account-001&goods_keyword=剪映&status=1", token)
	require.Equal(t, 0, list.Code)
	var listData struct {
		List []struct {
			ID           int64  `json:"id"`
			Account      string `json:"account"`
			GoodsKeyword string `json:"goods_keyword"`
			Reason       string `json:"reason"`
			Status       int    `json:"status"`
			StatusText   string `json:"status_text"`
			HitCount     int    `json:"hit_count"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(list.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, createData.ID, listData.List[0].ID)
	require.Equal(t, "risk-account-001", listData.List[0].Account)
	require.Equal(t, "剪映", listData.List[0].GoodsKeyword)
	require.Equal(t, "客户多次提交错误账号", listData.List[0].Reason)
	require.Equal(t, 1, listData.List[0].Status)
	require.Equal(t, "启用", listData.List[0].StatusText)
	require.Equal(t, 0, listData.List[0].HitCount)

	update := h.putJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID), map[string]any{
		"account":       "risk-account-001",
		"goods_keyword": "醒图",
		"reason":        "更新后的风控原因",
		"status":        0,
	}, token)
	require.Equal(t, 0, update.Code)

	disabled := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&goods_keyword=醒图&status=0", token)
	require.Equal(t, 0, disabled.Code)
	require.NoError(t, json.Unmarshal(disabled.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, "停用", listData.List[0].StatusText)

	enable := h.patchJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID)+"/status", map[string]any{"status": 1}, token)
	require.Equal(t, 0, enable.Code)

	deleted := h.deleteJSON("/api/admin/recharge-risks/rules/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, deleted.Code)

	empty := h.getJSON("/api/admin/recharge-risks/rules?page=1&page_size=20&account=risk-account-001", token)
	require.Equal(t, 0, empty.Code)
	require.NoError(t, json.Unmarshal(empty.Data, &listData))
	require.Empty(t, listData.List)
}

func TestRechargeRiskRuleSoftDeleteLongKeywordAllowsRecreate(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	keyword := strings.Repeat("k", 250)
	body := map[string]any{
		"account":       "long-keyword-account",
		"goods_keyword": keyword,
		"reason":        "长关键词删除后允许重建",
		"status":        1,
	}

	create := h.postJSON("/api/admin/recharge-risks/rules", body, token)
	require.Equal(t, 0, create.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &data))
	require.NotZero(t, data.ID)

	deleted := h.deleteJSON("/api/admin/recharge-risks/rules/"+int64ToString(data.ID), token)
	require.Equal(t, 0, deleted.Code)

	recreate := h.postJSON("/api/admin/recharge-risks/rules", body, token)
	require.Equal(t, 0, recreate.Code)
}

func TestRechargeRiskRuleDuplicateDBErrorMapsToConflict(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	_, err := h.app.Core().DB().Exec(context.Background(), `
CREATE TRIGGER recharge_risk_rule_duplicate_test
BEFORE INSERT ON recharge_risk_rule
FOR EACH ROW
BEGIN
    SIGNAL SQLSTATE '23000' SET MESSAGE_TEXT = 'Duplicate entry for recharge risk rule';
END
`)
	require.NoError(t, err)

	res := h.postJSON("/api/admin/recharge-risks/rules", map[string]any{
		"account":       "trigger-duplicate-account",
		"goods_keyword": "触发重复",
		"reason":        "数据库唯一冲突应返回 409",
		"status":        1,
	}, token)
	require.Equal(t, 409, res.Code)
}

func TestRechargeRiskRecordListFilters(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)
	now := h.app.Core().Now()
	_, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO recharge_risk_rule (
    account, goods_keyword, reason, status, hit_count,
    created_by_id, created_by_name, updated_by_id, updated_by_name,
    is_deleted, created_at, updated_at
) VALUES ('record-account-001', '微博', '记录筛选原因', 1, 1, 1, 'admin', 1, 'admin', 0, ?, ?)
`, now, now)
	require.NoError(t, err)
	ruleIDValue, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `SELECT id FROM recharge_risk_rule WHERE account = ?`, "record-account-001")
	require.NoError(t, err)
	ruleID := ruleIDValue.Int64()
	_, err = h.app.Core().DB().Exec(context.Background(), `
INSERT INTO recharge_risk_record (
    rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
    matched_keyword, reason, request_token_masked, intercepted_at, created_at
) VALUES (?, 1001, 'ORISKRECORD001', 'record-account-001', 2001, 'G-RISK-001', '新浪微博会员', '微博', '记录筛选原因', 'test***oken', ?, ?)
`, ruleID, now, now)
	require.NoError(t, err)

	start := url.QueryEscape(now.Add(-time.Minute).Format("2006-01-02 15:04:05"))
	end := url.QueryEscape(now.Add(time.Minute).Format("2006-01-02 15:04:05"))
	res := h.getJSON("/api/admin/recharge-risks/records?page=1&page_size=20&account=record-account-001&goods_keyword=微博&start_time="+start+"&end_time="+end, token)
	require.Equal(t, 0, res.Code)
	var data struct {
		List []struct {
			RuleID         int64  `json:"rule_id"`
			OrderNo        string `json:"order_no"`
			Account        string `json:"account"`
			MatchedKeyword string `json:"matched_keyword"`
			GoodsCode      string `json:"goods_code"`
			GoodsName      string `json:"goods_name"`
			Reason         string `json:"reason"`
			InterceptedAt  string `json:"intercepted_at"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.Len(t, data.List, 1)
	require.Equal(t, "ORISKRECORD001", data.List[0].OrderNo)
	require.Equal(t, "record-account-001", data.List[0].Account)
	require.Equal(t, "微博", data.List[0].MatchedKeyword)
	require.Equal(t, "G-RISK-001", data.List[0].GoodsCode)
	require.Equal(t, "新浪微博会员", data.List[0].GoodsName)
	require.Equal(t, "记录筛选原因", data.List[0].Reason)

	invalidTime := h.getJSON("/api/admin/recharge-risks/records?page=1&page_size=20&start_time=bad-time", token)
	require.Equal(t, 400, invalidTime.Code)
}
