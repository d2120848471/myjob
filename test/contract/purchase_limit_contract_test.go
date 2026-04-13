package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_PurchaseLimitPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/purchase-limit-strategies")
	require.Contains(t, res.body, "/api/admin/purchase-limit-strategies/enums")
	require.Contains(t, res.body, "/api/admin/purchase-limit-strategies/{id}/status")
}

func TestPurchaseLimitSeedsStayInSync(t *testing.T) {
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
`, 11)
	require.NoError(t, err)
	require.EqualValues(t, 11, menu.ID)
	require.Equal(t, "商品购买数量限制策略", menu.Name)
	require.Equal(t, "product.purchase_limit", menu.Code)
	require.Equal(t, 0, menu.SuperOnly)
	require.Equal(t, 11, menu.Sort)

	groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, 11)
	require.NoError(t, err)
	require.Equal(t, 1, groupMenuCount.Int())

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'商品购买数量限制策略'")
	require.Contains(t, string(seedFile), "'product.purchase_limit'")
}

func TestPurchaseLimitFlowsAndPermissions(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	adminMe := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, adminMe.Code)

	var adminMeData struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(adminMe.Data, &adminMeData))
	require.Contains(t, adminMeData.Permissions, "product.purchase_limit")

	menuTree := h.getJSON("/api/admin/menus/tree", token)
	require.Equal(t, 0, menuTree.Code)

	var menuTreeData struct {
		List []*menuTreeItem `json:"list"`
	}
	require.NoError(t, json.Unmarshal(menuTree.Data, &menuTreeData))
	menuCodes := flattenMenuCodes(menuTreeData.List)
	require.Contains(t, menuCodes, "product.purchase_limit")

	listBefore := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20", token)
	require.Equal(t, 0, listBefore.Code)

	var listBeforeData struct {
		List       []any `json:"list"`
		Pagination struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(listBefore.Data, &listBeforeData))
	require.Empty(t, listBeforeData.List)
	require.Equal(t, 1, listBeforeData.Pagination.Page)
	require.Equal(t, 20, listBeforeData.Pagination.PageSize)
	require.Equal(t, 0, listBeforeData.Pagination.Total)

	enumsRaw := h.rawRequest(http.MethodGet, "/api/admin/purchase-limit-strategies/enums", nil, token)
	require.Equal(t, http.StatusOK, enumsRaw.status)
	require.Equal(t, 0, enumsRaw.env.Code)

	var enumsData struct {
		LimitTypes []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"limit_types"`
		PeriodTypes []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"period_types"`
	}
	require.NoError(t, json.Unmarshal(enumsRaw.env.Data, &enumsData))
	require.Len(t, enumsData.LimitTypes, 2)
	require.Len(t, enumsData.PeriodTypes, 2)
	require.Equal(t, "同一会员", enumsData.LimitTypes[0].Title)
	require.Equal(t, "按区间(分钟)", enumsData.PeriodTypes[1].Title)

	invalidLimitType := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "非法限制类型",
		"limit_type":  9,
		"period_type": 1,
		"period":      1,
		"limit_nums":  1,
		"limit_times": 1,
	}, token)
	require.Equal(t, 400, invalidLimitType.Code)

	invalidPeriodType := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "非法周期类型",
		"limit_type":  1,
		"period_type": 9,
		"period":      1,
		"limit_nums":  1,
		"limit_times": 1,
	}, token)
	require.Equal(t, 400, invalidPeriodType.Code)

	invalidPeriod := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "非法周期",
		"limit_type":  1,
		"period_type": 1,
		"period":      0,
		"limit_nums":  1,
		"limit_times": 1,
	}, token)
	require.Equal(t, 400, invalidPeriod.Code)

	invalidLimitNums := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "非法数量",
		"limit_type":  1,
		"period_type": 1,
		"period":      1,
		"limit_nums":  -1,
		"limit_times": 1,
	}, token)
	require.Equal(t, 400, invalidLimitNums.Code)

	createRes := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "一天一号两次",
		"limit_type":  2,
		"period_type": 1,
		"period":      1,
		"limit_nums":  2,
		"limit_times": 2,
	}, token)
	require.Equal(t, 0, createRes.Code)

	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createRes.Data, &createData))
	require.NotZero(t, createData.ID)

	listAfterCreate := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20&keyword=一天一号", token)
	require.Equal(t, 0, listAfterCreate.Code)

	var listAfterCreateData struct {
		List []struct {
			ID              int64  `json:"id"`
			Name            string `json:"name"`
			LimitType       int    `json:"limit_type"`
			LimitTypeLabel  string `json:"limit_type_label"`
			PeriodType      int    `json:"period_type"`
			PeriodTypeLabel string `json:"period_type_label"`
			Period          int    `json:"period"`
			LimitNums       int    `json:"limit_nums"`
			LimitTimes      int    `json:"limit_times"`
			Status          int    `json:"status"`
			CreatedAt       string `json:"created_at"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listAfterCreate.Data, &listAfterCreateData))
	require.Len(t, listAfterCreateData.List, 1)
	require.Equal(t, createData.ID, listAfterCreateData.List[0].ID)
	require.Equal(t, "一天一号两次", listAfterCreateData.List[0].Name)
	require.Equal(t, 2, listAfterCreateData.List[0].LimitType)
	require.Equal(t, "同一充值账号", listAfterCreateData.List[0].LimitTypeLabel)
	require.Equal(t, 1, listAfterCreateData.List[0].PeriodType)
	require.Equal(t, "按天", listAfterCreateData.List[0].PeriodTypeLabel)
	require.Equal(t, 1, listAfterCreateData.List[0].Period)
	require.Equal(t, 2, listAfterCreateData.List[0].LimitNums)
	require.Equal(t, 2, listAfterCreateData.List[0].LimitTimes)
	require.Equal(t, 1, listAfterCreateData.List[0].Status)
	require.True(t, strings.TrimSpace(listAfterCreateData.List[0].CreatedAt) != "")

	createDefaultStatusRes := h.postJSON("/api/admin/purchase-limit-strategies", map[string]any{
		"name":        "默认开通策略",
		"limit_type":  1,
		"period_type": 1,
		"period":      1,
		"limit_nums":  0,
		"limit_times": 1,
	}, token)
	require.Equal(t, 0, createDefaultStatusRes.Code)

	var createDefaultStatusData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createDefaultStatusRes.Data, &createDefaultStatusData))
	require.NotZero(t, createDefaultStatusData.ID)

	defaultStatusList := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20&keyword=默认开通策略", token)
	require.Equal(t, 0, defaultStatusList.Code)

	var defaultStatusListData struct {
		List []struct {
			ID     int64 `json:"id"`
			Status int   `json:"status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(defaultStatusList.Data, &defaultStatusListData))
	require.Len(t, defaultStatusListData.List, 1)
	require.Equal(t, createDefaultStatusData.ID, defaultStatusListData.List[0].ID)
	require.Equal(t, 1, defaultStatusListData.List[0].Status)

	disableDefaultStatus := h.patchJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createDefaultStatusData.ID)+"/status", map[string]any{
		"status": 0,
	}, token)
	require.Equal(t, 0, disableDefaultStatus.Code)

	editKeepStatusRes := h.putJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createDefaultStatusData.ID), map[string]any{
		"name":        "默认开通策略-编辑",
		"limit_type":  1,
		"period_type": 2,
		"period":      15,
		"limit_nums":  3,
		"limit_times": 4,
	}, token)
	require.Equal(t, 0, editKeepStatusRes.Code)

	afterEditKeepStatus := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20&keyword=默认开通策略-编辑", token)
	require.Equal(t, 0, afterEditKeepStatus.Code)

	var afterEditKeepStatusData struct {
		List []struct {
			ID         int64 `json:"id"`
			Status     int   `json:"status"`
			PeriodType int   `json:"period_type"`
			Period     int   `json:"period"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterEditKeepStatus.Data, &afterEditKeepStatusData))
	require.Len(t, afterEditKeepStatusData.List, 1)
	require.Equal(t, createDefaultStatusData.ID, afterEditKeepStatusData.List[0].ID)
	require.Equal(t, 0, afterEditKeepStatusData.List[0].Status)
	require.Equal(t, 2, afterEditKeepStatusData.List[0].PeriodType)
	require.Equal(t, 15, afterEditKeepStatusData.List[0].Period)

	editRes := h.putJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createData.ID), map[string]any{
		"name":        "三分钟一号五次",
		"limit_type":  2,
		"period_type": 2,
		"period":      3,
		"limit_nums":  0,
		"limit_times": 5,
	}, token)
	require.Equal(t, 0, editRes.Code)

	statusRes := h.patchJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createData.ID)+"/status", map[string]any{
		"status": 0,
	}, token)
	require.Equal(t, 0, statusRes.Code)

	afterEdit := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20&keyword=三分钟一号", token)
	require.Equal(t, 0, afterEdit.Code)

	var afterEditData struct {
		List []struct {
			ID         int64 `json:"id"`
			PeriodType int   `json:"period_type"`
			Period     int   `json:"period"`
			LimitNums  int   `json:"limit_nums"`
			LimitTimes int   `json:"limit_times"`
			Status     int   `json:"status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterEdit.Data, &afterEditData))
	require.Len(t, afterEditData.List, 1)
	require.Equal(t, createData.ID, afterEditData.List[0].ID)
	require.Equal(t, 2, afterEditData.List[0].PeriodType)
	require.Equal(t, 3, afterEditData.List[0].Period)
	require.Equal(t, 0, afterEditData.List[0].LimitNums)
	require.Equal(t, 5, afterEditData.List[0].LimitTimes)
	require.Equal(t, 0, afterEditData.List[0].Status)

	deleteRes := h.deleteJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createData.ID), token)
	require.Equal(t, 0, deleteRes.Code)

	afterDelete := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20&keyword=三分钟一号", token)
	require.Equal(t, 0, afterDelete.Code)

	var afterDeleteData struct {
		List []any `json:"list"`
	}
	require.NoError(t, json.Unmarshal(afterDelete.Data, &afterDeleteData))
	require.Empty(t, afterDeleteData.List)

	deleteMissing := h.deleteJSON("/api/admin/purchase-limit-strategies/"+int64ToString(createData.ID), token)
	require.Equal(t, 400, deleteMissing.Code)

	limitedGroup := h.postJSON("/api/admin/groups", map[string]any{
		"name":        "限购观察组",
		"description": "没有限购策略权限",
	}, token)
	require.Equal(t, 0, limitedGroup.Code)

	var limitedGroupData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(limitedGroup.Data, &limitedGroupData))
	require.NotZero(t, limitedGroupData.ID)

	createLimitedUser := h.postJSON("/api/admin/users", map[string]any{
		"username":         "limit01",
		"confirm_username": "limit01",
		"password":         "Limit_123",
		"confirm_password": "Limit_123",
		"real_name":        "Limit",
		"phone":            "13800003333",
		"group_id":         limitedGroupData.ID,
	}, token)
	require.Equal(t, 0, createLimitedUser.Code)

	limitedLogin := h.postJSON("/api/admin/auth/login", map[string]any{
		"username": "limit01",
		"password": "Limit_123",
	}, "")
	require.Equal(t, 0, limitedLogin.Code)

	var limitedLoginData struct {
		Token         string `json:"token"`
		NeedSMSVerify bool   `json:"need_sms_verify"`
		LoginToken    string `json:"login_token"`
	}
	require.NoError(t, json.Unmarshal(limitedLogin.Data, &limitedLoginData))
	if limitedLoginData.NeedSMSVerify {
		sendRes := h.postJSON("/api/admin/auth/sms/send", map[string]any{
			"login_token": limitedLoginData.LoginToken,
		}, "")
		require.Equal(t, 0, sendRes.Code)
		code := h.lastSMSCode(t, "13800003333")

		verifyRes := h.postJSON("/api/admin/auth/sms/verify", map[string]any{
			"login_token": limitedLoginData.LoginToken,
			"sms_code":    code,
		}, "")
		require.Equal(t, 0, verifyRes.Code)

		var verifyData struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(verifyRes.Data, &verifyData))
		limitedLoginData.Token = verifyData.Token
	}
	require.NotEmpty(t, limitedLoginData.Token)

	forbiddenRes := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20", limitedLoginData.Token)
	require.Equal(t, 403, forbiddenRes.Code)

	saveLimitedAuth := h.patchJSON("/api/admin/groups/"+int64ToString(limitedGroupData.ID)+"/permissions", map[string]any{
		"menu_ids": []int64{11},
	}, token)
	require.Equal(t, 0, saveLimitedAuth.Code)

	allowedRes := h.getJSON("/api/admin/purchase-limit-strategies?page=1&page_size=20", limitedLoginData.Token)
	require.Equal(t, 0, allowedRes.Code)

	var logCount struct {
		Count int `db:"count"`
	}
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &logCount, `
SELECT COUNT(*) AS count
FROM admin_operation_log
WHERE description LIKE ?
`, "%商品购买数量限制策略%")
	require.NoError(t, err)
	require.GreaterOrEqual(t, logCount.Count, 3)
}

type menuTreeItem struct {
	ID       int64           `json:"id"`
	Code     string          `json:"code"`
	Children []*menuTreeItem `json:"children"`
}

func flattenMenuCodes(items []*menuTreeItem) []string {
	result := make([]string, 0)
	var walk func(list []*menuTreeItem)
	walk = func(list []*menuTreeItem) {
		for _, item := range list {
			if item == nil {
				continue
			}
			if strings.TrimSpace(item.Code) != "" {
				result = append(result, item.Code)
			}
			walk(item.Children)
		}
	}
	walk(items)
	return result
}

func int64ToString(v int64) string {
	return strconv.FormatInt(v, 10)
}
