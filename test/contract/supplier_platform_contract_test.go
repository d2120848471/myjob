package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_SupplierPlatformPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/supplier-platform-types")
	require.Contains(t, res.body, "/api/admin/supplier-platforms")
	require.Contains(t, res.body, "/api/admin/supplier-platforms/{id}")
	require.Contains(t, res.body, "/api/admin/supplier-platforms/{id}/balance/refresh")
}

func TestSupplierPlatformSeedsStayInSync(t *testing.T) {
	h := newTestHarness(t)

	var menu struct {
		ID        int64  `db:"id"`
		Name      string `db:"name"`
		Code      string `db:"code"`
		SuperOnly int    `db:"super_only"`
		Sort      int    `db:"sort"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &menu, `
SELECT id, name, code, super_only, sort
FROM admin_menu
WHERE id = ?
`, 12)
	require.NoError(t, err)
	require.EqualValues(t, 12, menu.ID)
	require.Equal(t, "第三方对接", menu.Name)
	require.Equal(t, "supplier.index", menu.Code)
	require.Equal(t, 0, menu.SuperOnly)
	require.Equal(t, 12, menu.Sort)

	groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, 12)
	require.NoError(t, err)
	require.Equal(t, 1, groupMenuCount.Int())

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'第三方对接'")
	require.Contains(t, string(seedFile), "'supplier.index'")
}

func TestSupplierPlatformFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	adminMe := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, adminMe.Code)

	var adminMeData struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(adminMe.Data, &adminMeData))
	require.Contains(t, adminMeData.Permissions, "supplier.index")

	typesRes := h.getJSON("/api/admin/supplier-platform-types", token)
	require.Equal(t, 0, typesRes.Code)

	var typeData struct {
		List []struct {
			ID           int    `json:"id"`
			TypeName     string `json:"type_name"`
			ProviderCode string `json:"provider_code"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(typesRes.Data, &typeData))
	require.Len(t, typeData.List, 8)
	require.Contains(t, typeData.List, struct {
		ID           int    `json:"id"`
		TypeName     string `json:"type_name"`
		ProviderCode string `json:"provider_code"`
	}{ID: 35, TypeName: "星权益", ProviderCode: "xingquanyi"})

	createSubject := h.postJSON("/api/admin/subjects", map[string]any{
		"name":    "供货主体A",
		"has_tax": 1,
	}, token)
	require.Equal(t, 0, createSubject.Code)

	var subjectData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createSubject.Data, &subjectData))
	require.NotZero(t, subjectData.ID)

	invalidDomain := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             "非法域名平台",
		"domain":           "https://xqy.test.local",
		"backup_domain":    "xqy.test.local",
		"type_id":          35,
		"subject_id":       subjectData.ID,
		"has_tax":          1,
		"token_id":         "1008612345",
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
	}, token)
	require.Equal(t, 400, invalidDomain.Code)

	mode := "success"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/customer", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		if mode == "success" {
			_, _ = w.Write([]byte(`{"code":"ok","msg":"","data":{"balance":"24588.5010"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":"invalid_sign","msg":"签名错误"}`))
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	createRes := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             "木木（星权益含税）",
		"domain":           host,
		"backup_domain":    host,
		"type_id":          35,
		"subject_id":       subjectData.ID,
		"has_tax":          1,
		"token_id":         "1008612345",
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
	}, token)
	require.Equal(t, 0, createRes.Code)

	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createRes.Data, &createData))
	require.NotZero(t, createData.ID)

	detailRes := h.getJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, detailRes.Code)

	var detailData struct {
		ID              int64                  `json:"id"`
		Name            string                 `json:"name"`
		Domain          string                 `json:"domain"`
		BackupDomain    string                 `json:"backup_domain"`
		TypeID          int                    `json:"type_id"`
		SubjectID       int64                  `json:"subject_id"`
		HasTax          int                    `json:"has_tax"`
		TokenID         string                 `json:"token_id"`
		SecretKey       string                 `json:"secret_key"`
		ProviderCode    string                 `json:"provider_code"`
		ProviderName    string                 `json:"provider_name"`
		ThresholdAmount string                 `json:"threshold_amount"`
		ExtraConfig     map[string]interface{} `json:"extra_config"`
	}
	require.NoError(t, json.Unmarshal(detailRes.Data, &detailData))
	require.Equal(t, createData.ID, detailData.ID)
	require.Equal(t, "secret-key", detailData.SecretKey)
	require.Equal(t, "xingquanyi", detailData.ProviderCode)
	require.Equal(t, "星权益", detailData.ProviderName)
	require.Equal(t, "5000.0000", detailData.ThresholdAmount)
	require.NotNil(t, detailData.ExtraConfig)

	listRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=木木", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			ID                int64  `json:"id"`
			ProviderCode      string `json:"provider_code"`
			ProviderName      string `json:"provider_name"`
			TypeName          string `json:"type_name"`
			SubjectName       string `json:"subject_name"`
			ConnectStatus     int    `json:"connect_status"`
			ConnectStatusText string `json:"connect_status_text"`
			BalanceWarning    int    `json:"balance_warning"`
			LastBalance       string `json:"last_balance"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, "xingquanyi", listData.List[0].ProviderCode)
	require.Equal(t, "星权益", listData.List[0].ProviderName)
	require.Equal(t, "星权益", listData.List[0].TypeName)
	require.Equal(t, "供货主体A", listData.List[0].SubjectName)
	require.Equal(t, 0, listData.List[0].ConnectStatus)
	require.Equal(t, "未验证", listData.List[0].ConnectStatusText)
	require.Equal(t, 0, listData.List[0].BalanceWarning)
	require.Equal(t, "", listData.List[0].LastBalance)

	refreshSuccess := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, refreshSuccess.Code)

	var refreshSuccessData struct {
		ID                int64  `json:"id"`
		Balance           string `json:"balance"`
		ConnectStatus     int    `json:"connect_status"`
		ConnectStatusText string `json:"connect_status_text"`
		Message           string `json:"message"`
		TraceID           string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(refreshSuccess.Data, &refreshSuccessData))
	require.Equal(t, createData.ID, refreshSuccessData.ID)
	require.Equal(t, "24588.5010", refreshSuccessData.Balance)
	require.Equal(t, 1, refreshSuccessData.ConnectStatus)
	require.Equal(t, "正常", refreshSuccessData.ConnectStatusText)
	require.Equal(t, "查询成功", refreshSuccessData.Message)
	require.NotEmpty(t, refreshSuccessData.TraceID)

	updateRes := h.putJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID), map[string]any{
		"name":             "木木（星权益含税）-编辑",
		"domain":           host,
		"backup_domain":    host,
		"type_id":          35,
		"subject_id":       subjectData.ID,
		"has_tax":          1,
		"token_id":         "10086",
		"secret_key":       "secret-key-updated",
		"threshold_amount": "30000.0000",
		"sort":             1,
		"crowd_name":       "运营群-编辑",
	}, token)
	require.Equal(t, 0, updateRes.Code)

	afterReset := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=编辑", token)
	require.Equal(t, 0, afterReset.Code)

	var afterResetData struct {
		List []struct {
			ConnectStatus     int    `json:"connect_status"`
			ConnectStatusText string `json:"connect_status_text"`
			LastBalance       string `json:"last_balance"`
			BalanceWarning    int    `json:"balance_warning"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterReset.Data, &afterResetData))
	require.Len(t, afterResetData.List, 1)
	require.Equal(t, 0, afterResetData.List[0].ConnectStatus)
	require.Equal(t, "未验证", afterResetData.List[0].ConnectStatusText)
	require.Equal(t, "24588.5010", afterResetData.List[0].LastBalance)
	require.Equal(t, 1, afterResetData.List[0].BalanceWarning)

	mode = "fail"
	refreshFail := h.postJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID)+"/balance/refresh", map[string]any{}, token)
	require.Equal(t, 0, refreshFail.Code)

	var refreshFailData struct {
		Balance           string `json:"balance"`
		ConnectStatus     int    `json:"connect_status"`
		ConnectStatusText string `json:"connect_status_text"`
		Message           string `json:"message"`
		TraceID           string `json:"trace_id"`
	}
	require.NoError(t, json.Unmarshal(refreshFail.Data, &refreshFailData))
	require.Equal(t, "24588.5010", refreshFailData.Balance)
	require.Equal(t, 2, refreshFailData.ConnectStatus)
	require.Equal(t, "异常", refreshFailData.ConnectStatusText)
	require.Equal(t, "签名错误", refreshFailData.Message)
	require.NotEmpty(t, refreshFailData.TraceID)

	var account struct {
		IsDeleted          int    `db:"is_deleted"`
		LastBalance        string `db:"last_balance"`
		LastBalanceStatus  int    `db:"last_balance_status"`
		LastBalanceMessage string `db:"last_balance_message"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &account, `
SELECT is_deleted, last_balance, last_balance_status, last_balance_message
FROM supplier_platform_account
WHERE id = ?
`, createData.ID)
	require.NoError(t, err)
	require.Equal(t, 0, account.IsDeleted)
	require.Equal(t, "24588.5010", account.LastBalance)
	require.Equal(t, 2, account.LastBalanceStatus)
	require.Equal(t, "签名错误", account.LastBalanceMessage)

	var balanceLog struct {
		Count            int    `db:"count"`
		RequestSnapshot  string `db:"request_snapshot"`
		ResponseSnapshot string `db:"response_snapshot"`
	}
	err = h.app.Core().DB().GetCore().GetScan(context.Background(), &balanceLog, `
SELECT
    COUNT(*) AS count,
    MAX(request_snapshot) AS request_snapshot,
    MAX(CASE WHEN success = 0 THEN response_snapshot END) AS response_snapshot
FROM supplier_platform_balance_log
WHERE platform_id = ?
`, createData.ID)
	require.NoError(t, err)
	require.Equal(t, 2, balanceLog.Count)
	require.NotContains(t, balanceLog.RequestSnapshot, "secret-key-updated")
	require.NotContains(t, balanceLog.RequestSnapshot, "1008612345")
	require.Contains(t, balanceLog.RequestSnapshot, "1008")
	require.Contains(t, balanceLog.RequestSnapshot, "2345")
	require.Contains(t, balanceLog.ResponseSnapshot, "invalid_sign")

	deleteRes := h.deleteJSON("/api/admin/supplier-platforms/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, deleteRes.Code)

	afterDelete := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=编辑", token)
	require.Equal(t, 0, afterDelete.Code)

	var afterDeleteData struct {
		List []any `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterDelete.Data, &afterDeleteData))
	require.Empty(t, afterDeleteData.List)

	err = h.app.Core().DB().GetCore().GetScan(context.Background(), &account, `
	SELECT is_deleted, last_balance, last_balance_status, last_balance_message
	FROM supplier_platform_account
	WHERE id = ?
	`, createData.ID)
	require.NoError(t, err)
	require.Equal(t, 1, account.IsDeleted)

	recreateRes := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             "木木（星权益含税）-重建",
		"domain":           host,
		"backup_domain":    host,
		"type_id":          35,
		"subject_id":       subjectData.ID,
		"has_tax":          1,
		"token_id":         "10086",
		"secret_key":       "secret-key-updated",
		"threshold_amount": "30000.0000",
		"sort":             9,
		"crowd_name":       "运营群-重建",
	}, token)
	require.Equal(t, 0, recreateRes.Code)

	var recreateData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(recreateRes.Data, &recreateData))
	require.NotZero(t, recreateData.ID)

	secondDeleteRes := h.deleteJSON("/api/admin/supplier-platforms/"+int64ToString(recreateData.ID), token)
	require.Equal(t, 0, secondDeleteRes.Code)

	rows, err := h.app.Core().DB().GetCore().GetAll(context.Background(), `
	SELECT id, token_id
	FROM supplier_platform_account
	WHERE subject_id = ? AND is_deleted = 1
	ORDER BY id
	`, subjectData.ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	firstTokenID := rows[0]["token_id"].String()
	secondTokenID := rows[1]["token_id"].String()
	require.NotEqual(t, firstTokenID, secondTokenID)
	require.NotEqual(t, "10086", firstTokenID)
	require.NotEqual(t, "10086", secondTokenID)
	require.Contains(t, firstTokenID, "__deleted__")
	require.Contains(t, secondTokenID, "__deleted__")
}

func TestSupplierPlatformStatus_DisableCascadesBindingsAndReenableKeepsBindingsDisabled(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "平台状态联动", "视频充值", "会员周卡")
	platformSubjectID := h.createSubject(t, token, "平台状态主体", 0)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "平台状态商品", "18.8000")
	platformID := h.createSupplierPlatformAccount(t, token, "平台状态账号", platformSubjectID, 0, "platform-status-account")

	createBindingRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "SKU-STATUS-1",
		"supplier_goods_name": "平台状态渠道商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 0, createBindingRes.Code)

	detailBeforeDisable := h.getJSON("/api/admin/supplier-platforms/"+int64ToString(platformID), token)
	require.Equal(t, 0, detailBeforeDisable.Code)

	var detailBeforeDisableData struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.Unmarshal(detailBeforeDisable.Data, &detailBeforeDisableData))
	require.Equal(t, 1, detailBeforeDisableData.Status)

	listBeforeDisable := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台状态账号", token)
	require.Equal(t, 0, listBeforeDisable.Code)

	var listBeforeDisableData struct {
		List []struct {
			ID     int64 `json:"id"`
			Status int   `json:"status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listBeforeDisable.Data, &listBeforeDisableData))
	require.Len(t, listBeforeDisableData.List, 1)
	require.Equal(t, platformID, listBeforeDisableData.List[0].ID)
	require.Equal(t, 1, listBeforeDisableData.List[0].Status)

	h.updateSupplierPlatformStatus(t, token, platformID, "平台状态账号", platformSubjectID, 0, "platform-status-account", 0)

	var platformAccount struct {
		Status int `db:"status"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &platformAccount, `
SELECT status
FROM supplier_platform_account
WHERE id = ?
`, platformID)
	require.NoError(t, err)
	require.Equal(t, 0, platformAccount.Status)

	var binding struct {
		DockStatus int `db:"dock_status"`
	}
	err = h.app.Core().DB().GetCore().GetScan(context.Background(), &binding, `
SELECT dock_status
FROM product_goods_channel_binding
WHERE goods_id = ? AND platform_account_id = ? AND is_deleted = 0
`, goodsID, platformID)
	require.NoError(t, err)
	require.Equal(t, 0, binding.DockStatus)

	listAfterDisablePlatform := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台状态账号", token)
	require.Equal(t, 0, listAfterDisablePlatform.Code)

	var listAfterDisablePlatformData struct {
		List []struct {
			ID     int64 `json:"id"`
			Status int   `json:"status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listAfterDisablePlatform.Data, &listAfterDisablePlatformData))
	require.Len(t, listAfterDisablePlatformData.List, 1)
	require.Equal(t, platformID, listAfterDisablePlatformData.List[0].ID)
	require.Equal(t, 0, listAfterDisablePlatformData.List[0].Status)

	listAfterDisable := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=平台状态商品", token)
	require.Equal(t, 0, listAfterDisable.Code)

	var listAfterDisableData struct {
		List []struct {
			ID                int64    `json:"id"`
			BoundChannels     []string `json:"bound_channels"`
			BoundChannelCount int      `json:"bound_channel_count"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listAfterDisable.Data, &listAfterDisableData))
	require.Len(t, listAfterDisableData.List, 1)
	require.Equal(t, goodsID, listAfterDisableData.List[0].ID)
	require.Empty(t, listAfterDisableData.List[0].BoundChannels)
	require.Equal(t, 0, listAfterDisableData.List[0].BoundChannelCount)

	h.updateSupplierPlatformStatus(t, token, platformID, "平台状态账号", platformSubjectID, 0, "platform-status-account", 1)

	detailAfterEnable := h.getJSON("/api/admin/supplier-platforms/"+int64ToString(platformID), token)
	require.Equal(t, 0, detailAfterEnable.Code)

	var detailAfterEnableData struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.Unmarshal(detailAfterEnable.Data, &detailAfterEnableData))
	require.Equal(t, 1, detailAfterEnableData.Status)

	listAfterEnablePlatform := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台状态账号", token)
	require.Equal(t, 0, listAfterEnablePlatform.Code)

	var listAfterEnablePlatformData struct {
		List []struct {
			ID     int64 `json:"id"`
			Status int   `json:"status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listAfterEnablePlatform.Data, &listAfterEnablePlatformData))
	require.Len(t, listAfterEnablePlatformData.List, 1)
	require.Equal(t, platformID, listAfterEnablePlatformData.List[0].ID)
	require.Equal(t, 1, listAfterEnablePlatformData.List[0].Status)

	err = h.app.Core().DB().GetCore().GetScan(context.Background(), &binding, `
SELECT dock_status
FROM product_goods_channel_binding
WHERE goods_id = ? AND platform_account_id = ? AND is_deleted = 0
`, goodsID, platformID)
	require.NoError(t, err)
	require.Equal(t, 0, binding.DockStatus)

	listAfterEnable := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=平台状态商品", token)
	require.Equal(t, 0, listAfterEnable.Code)
	require.NoError(t, json.Unmarshal(listAfterEnable.Data, &listAfterDisableData))
	require.Len(t, listAfterDisableData.List, 1)
	require.Empty(t, listAfterDisableData.List[0].BoundChannels)
	require.Equal(t, 0, listAfterDisableData.List[0].BoundChannelCount)
}

func TestSupplierPlatformList_StatusFilter(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	platformSubjectID := h.createSubject(t, token, "平台筛选主体", 0)
	enabledPlatformID := h.createSupplierPlatformAccount(t, token, "平台筛选账号", platformSubjectID, 0, "status-filter-enabled")
	disabledPlatformID := h.createSupplierPlatformAccount(t, token, "平台筛选账号", platformSubjectID, 0, "status-filter-disabled")
	h.updateSupplierPlatformStatus(t, token, disabledPlatformID, "平台筛选账号", platformSubjectID, 0, "status-filter-disabled", 0)

	assertList := func(res apiEnvelope, expectedIDs ...int64) {
		t.Helper()
		require.Equal(t, 0, res.Code)

		var data struct {
			List []struct {
				ID     int64 `json:"id"`
				Status int   `json:"status"`
			} `json:"list"`
		}
		require.NoError(t, json.Unmarshal(res.Data, &data))
		require.Len(t, data.List, len(expectedIDs))
		for index, expectedID := range expectedIDs {
			require.Equal(t, expectedID, data.List[index].ID)
		}
	}

	allRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台筛选账号", token)
	assertList(allRes, disabledPlatformID, enabledPlatformID)

	allWithSentinelRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台筛选账号&status=-1", token)
	assertList(allWithSentinelRes, disabledPlatformID, enabledPlatformID)

	enabledOnlyRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台筛选账号&status=1", token)
	assertList(enabledOnlyRes, enabledPlatformID)

	disabledOnlyRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台筛选账号&status=0", token)
	assertList(disabledOnlyRes, disabledPlatformID)

	invalidStatusRes := h.getJSON("/api/admin/supplier-platforms?page=1&page_size=20&keyword=平台筛选账号&status=2", token)
	require.Equal(t, 400, invalidStatusRes.Code)
	require.Equal(t, "平台状态筛选值错误", invalidStatusRes.Message)

	detailEnabledRes := h.getJSON("/api/admin/supplier-platforms/"+int64ToString(enabledPlatformID), token)
	require.Equal(t, 0, detailEnabledRes.Code)
}
