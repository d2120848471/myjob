package contract_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomerSchemaAndPermissionSeed(t *testing.T) {
	h := newTestHarness(t)

	tableCount, err := h.app.Core().DB().GetCore().GetValue(
		context.Background(),
		`SELECT COUNT(*) FROM customer_user`,
	)
	require.NoError(t, err)
	require.Equal(t, 0, tableCount.Int())

	var menu struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Code string `db:"code"`
	}
	err = h.app.Core().DB().GetCore().GetScan(
		context.Background(),
		&menu,
		`SELECT id, name, code FROM admin_menu WHERE code = ?`,
		"customer.manage",
	)
	require.NoError(t, err)
	require.NotZero(t, menu.ID)
	require.Equal(t, "客户管理", menu.Name)

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'客户管理'")
	require.Contains(t, string(seedFile), "'customer.manage'")
}

func TestAdminCustomerManagementFlow(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	create := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "后台新增客户",
		"phone":                "13800008888",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, token)
	require.Equal(t, 0, create.Code)
	var createData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &createData))
	require.NotZero(t, createData.ID)

	list := h.getJSON("/api/admin/customers?page=1&page_size=20&keyword=后台新增客户", token)
	require.Equal(t, 0, list.Code)
	require.Contains(t, string(list.Data), "后台新增客户")

	detail := h.getJSON(fmt.Sprintf("/api/admin/customers/%d", createData.ID), token)
	require.Equal(t, 0, detail.Code)

	edit := h.putJSON(fmt.Sprintf("/api/admin/customers/%d", createData.ID), map[string]any{
		"company_name": "后台编辑客户",
		"phone":        "13800009999",
		"status":       1,
	}, token)
	require.Equal(t, 0, edit.Code)

	disable := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/status", createData.ID), map[string]any{
		"status": 0,
	}, token)
	require.Equal(t, 0, disable.Code)

	resetPassword := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/password", createData.ID), map[string]any{
		"password":         "New_123",
		"confirm_password": "New_123",
	}, token)
	require.Equal(t, 0, resetPassword.Code)
	require.NotContains(t, string(resetPassword.Data), "New_123")

	resetPayPassword := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/pay-password", createData.ID), map[string]any{
		"pay_password":         "654321",
		"confirm_pay_password": "654321",
	}, token)
	require.Equal(t, 0, resetPayPassword.Code)
	require.NotContains(t, string(resetPayPassword.Data), "654321")

	deleteRes := h.deleteJSON(fmt.Sprintf("/api/admin/customers/%d", createData.ID), token)
	require.Equal(t, 0, deleteRes.Code)

	trash := h.getJSON("/api/admin/customers/trash?page=1&page_size=20&keyword=后台编辑客户", token)
	require.Equal(t, 0, trash.Code)
	require.Contains(t, string(trash.Data), "后台编辑客户")

	restore := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/restore", createData.ID), map[string]any{}, token)
	require.Equal(t, 0, restore.Code)

	rows := make([]struct {
		Description string `db:"description"`
	}, 0)
	err := h.app.Core().DB().GetCore().GetScan(context.Background(), &rows, `
SELECT description
FROM admin_operation_log
WHERE description LIKE '切换客户状态%'
   OR description LIKE '删除客户%'
   OR description LIKE '恢复客户%'
   OR description LIKE '重置客户%'
ORDER BY id
`)
	require.NoError(t, err)
	require.NotEmpty(t, rows)
	for _, row := range rows {
		require.Contains(t, row.Description, fmt.Sprintf("客户ID=%d", createData.ID))
		require.NotContains(t, row.Description, "13800009999")
		require.NotContains(t, row.Description, "New_123")
		require.NotContains(t, row.Description, "654321")
	}
}

func TestCustomerPhoneStaysReservedInTrash(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	create := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "手机号占用客户",
		"phone":                "13800010001",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, token)
	require.Equal(t, 0, create.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &data))

	require.Equal(t, 0, h.deleteJSON(fmt.Sprintf("/api/admin/customers/%d", data.ID), token).Code)

	duplicate := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "重复手机号客户",
		"phone":                "13800010001",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, token)
	require.NotEqual(t, 0, duplicate.Code)
}

func TestCustomerTrashBlocksNormalWriteActions(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	create := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "回收站保护客户",
		"phone":                "13800010003",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, token)
	require.Equal(t, 0, create.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &data))

	require.Equal(t, 0, h.deleteJSON(fmt.Sprintf("/api/admin/customers/%d", data.ID), token).Code)

	edit := h.putJSON(fmt.Sprintf("/api/admin/customers/%d", data.ID), map[string]any{
		"company_name": "不应编辑",
		"phone":        "13800010004",
		"status":       1,
	}, token)
	require.NotEqual(t, 0, edit.Code)
	require.NotEqual(t, 0, h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/status", data.ID), map[string]any{"status": 1}, token).Code)
	require.NotEqual(t, 0, h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/password", data.ID), map[string]any{
		"password":         "New_123",
		"confirm_password": "New_123",
	}, token).Code)
	require.NotEqual(t, 0, h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/pay-password", data.ID), map[string]any{
		"pay_password":         "654321",
		"confirm_pay_password": "654321",
	}, token).Code)

	duplicateOldPhone := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "旧手机号仍占用",
		"phone":                "13800010003",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, token)
	require.NotEqual(t, 0, duplicateOldPhone.Code)
}
