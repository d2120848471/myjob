package contract_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductGoodsChannelPriceChangePermissionSeeded(t *testing.T) {
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
`, 16)
	require.NoError(t, err)
	require.EqualValues(t, 16, menu.ID)
	require.Equal(t, "自动改价记录", menu.Name)
	require.Equal(t, "product.price_change", menu.Code)
	require.Equal(t, 0, menu.SuperOnly)
	require.Equal(t, 16, menu.Sort)

	groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, 16)
	require.NoError(t, err)
	require.Equal(t, 1, groupMenuCount.Int())

	token := h.loginAdmin(t)
	me := h.getJSON("/api/admin/auth/me", token)
	require.Equal(t, 0, me.Code)
	var meData struct {
		Permissions []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(me.Data, &meData))
	require.Contains(t, meData.Permissions, "product.price_change")

	menuTree := h.getJSON("/api/admin/menus/tree", token)
	require.Equal(t, 0, menuTree.Code)
	var menuTreeData struct {
		List []*menuTreeItem `json:"list"`
	}
	require.NoError(t, json.Unmarshal(menuTree.Data, &menuTreeData))
	require.Contains(t, flattenMenuCodes(menuTreeData.List), "product.price_change")

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'自动改价记录'")
	require.Contains(t, string(seedFile), "'product.price_change'")
	require.Contains(t, string(seedFile), "(1, 16, NOW())")
}

func TestProductGoodsChannelPriceChangeList(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	h.seedProductGoodsChannelPriceChange(t, "push", "PRICE-CHANGE-001", "2582531")

	list := h.getJSON("/api/admin/product-goods-channel-price-changes?page=1&page_size=20&source=push&keyword=PRICE-CHANGE-001", token)
	require.Equal(t, 0, list.Code)
	require.Contains(t, string(list.Data), "PRICE-CHANGE-001")
	require.Contains(t, string(list.Data), "2582531")
	require.Contains(t, string(list.Data), "push")
	require.Contains(t, string(list.Data), "变动前")
}

func TestProductGoodsChannelPriceChangeRequiresDedicatedPermission(t *testing.T) {
	h := newTestHarness(t)
	adminToken := h.loginAdmin(t)
	h.seedProductGoodsChannelPriceChange(t, "monitor", "PRICE-CHANGE-PERM", "2582532")

	productGoodsToken := h.createLimitedUserToken(t, adminToken, 13)
	productGoodsRes := h.getJSON("/api/admin/product-goods-channel-price-changes?page=1&page_size=20", productGoodsToken)
	require.Equal(t, 403, productGoodsRes.Code)

	priceChangeToken := h.createLimitedUserToken(t, adminToken, 16)
	priceChangeRes := h.getJSON("/api/admin/product-goods-channel-price-changes?page=1&page_size=20&keyword=PRICE-CHANGE-PERM", priceChangeToken)
	require.Equal(t, 0, priceChangeRes.Code)
	require.Contains(t, string(priceChangeRes.Data), "PRICE-CHANGE-PERM")
}

func (h *testHarness) seedProductGoodsChannelPriceChange(t *testing.T, source, goodsCode, supplierGoodsNo string) int64 {
	t.Helper()
	now := h.app.Core().Now()
	result, err := h.app.Core().DB().Exec(context.Background(), `
INSERT INTO product_goods_channel_price_change_log (
    source, provider_code, platform_account_id, platform_account_name, binding_id,
    goods_id, goods_code, goods_name, goods_icon, supplier_goods_no, supplier_goods_name,
    old_source_cost_price, new_source_cost_price, old_cost_price, new_cost_price,
    old_effective_sell_price, new_effective_sell_price, change_amount,
    description, raw_payload, changed_at, created_at
) VALUES (?, 'kakayun', 1, '卡卡云测试账号', 1, 1, ?, '自动改价测试商品', '', ?, '上游测试商品', '10.0000', '12.0000', '10.0000', '12.0000', '20.0000', '22.0000', '2.0000', '变动前 10.0000，变动后 12.0000', '{}', ?, ?)
`, source, goodsCode, supplierGoodsNo, now, now)
	require.NoError(t, err)
	id, err := result.LastInsertId()
	require.NoError(t, err)
	return id
}
