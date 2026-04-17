package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_ProductGoodsChannelConfigPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-config")
}

func TestProductGoodsChannelConfigSeedsStayInSync(t *testing.T) {
	h := newTestHarness(t)

	type menuRow struct {
		ID       int64  `db:"id"`
		ParentID int64  `db:"parent_id"`
		Name     string `db:"name"`
		Code     string `db:"code"`
		Sort     int    `db:"sort"`
	}

	for _, item := range []struct {
		ID       int64
		Name     string
		Code     string
		ParentID int64
	}{
		{ID: 14, ParentID: 13, Name: "商品渠道配置查看", Code: "product.goods.channel.view"},
		{ID: 15, ParentID: 13, Name: "商品渠道配置编辑", Code: "product.goods.channel.edit"},
	} {
		var row menuRow
		err := h.app.Core().DB().GetCore().GetScan(context.Background(), &row, `
SELECT id, parent_id, name, code, sort
FROM admin_menu
WHERE id = ?
`, item.ID)
		require.NoError(t, err)
		require.EqualValues(t, item.ID, row.ID)
		require.EqualValues(t, item.ParentID, row.ParentID)
		require.Equal(t, item.Name, row.Name)
		require.Equal(t, item.Code, row.Code)

		groupMenuCount, err := h.app.Core().DB().GetCore().GetValue(context.Background(), `
SELECT COUNT(*)
FROM admin_group_menu
WHERE group_id = 1 AND menu_id = ?
`, item.ID)
		require.NoError(t, err)
		require.Equal(t, 1, groupMenuCount.Int())
	}

	seedFile, err := os.ReadFile(filepath.Join("..", "..", "manifest", "sql", "002_seed_menu.sql"))
	require.NoError(t, err)
	require.Contains(t, string(seedFile), "'商品渠道配置查看'")
	require.Contains(t, string(seedFile), "'product.goods.channel.view'")
	require.Contains(t, string(seedFile), "'商品渠道配置编辑'")
	require.Contains(t, string(seedFile), "'product.goods.channel.edit'")
}

func TestProductGoodsChannelConfig_GetAutoInitAndUpdate(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	ctx := context.Background()
	now := time.Now()

	goodsCode := "T-CHCFG-001"
	createGoodsResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, created_at, updated_at)
VALUES (?, 1, '渠道配置测试商品', 'card_secret', ?, ?)
`, goodsCode, now, now)
	require.NoError(t, err)
	goodsID, err := createGoodsResult.LastInsertId()
	require.NoError(t, err)

	getDefaultRaw := h.rawRequest(http.MethodGet, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-config", nil, token)
	require.Equal(t, http.StatusOK, getDefaultRaw.status)
	getDefault := getDefaultRaw.env
	require.Equal(t, 0, getDefault.Code)
	var getDefaultData struct {
		GoodsID                int64   `json:"goods_id"`
		GoodsCode              string  `json:"goods_code"`
		GoodsName              string  `json:"goods_name"`
		RouteMode              string  `json:"route_mode"`
		AttemptTimeoutEnabled  bool    `json:"attempt_timeout_enabled"`
		AttemptTimeoutMinutes  int     `json:"attempt_timeout_minutes"`
		AllowLoss              bool    `json:"allow_loss"`
		MaxLossAmount          *string `json:"max_loss_amount"`
		BoundChannelCount      int     `json:"bound_channel_count"`
		PrimaryChannelName     string  `json:"primary_channel_name"`
		MinChannelCost         string  `json:"min_channel_cost"`
		ChannelAutoPriceStatus bool    `json:"channel_auto_price_status"`
	} // 只断言本模块关心的字段
	require.NoError(t, json.Unmarshal(getDefault.Data, &getDefaultData))
	require.Equal(t, goodsID, getDefaultData.GoodsID)
	require.Equal(t, goodsCode, getDefaultData.GoodsCode)
	require.Equal(t, "渠道配置测试商品", getDefaultData.GoodsName)
	require.Equal(t, "fixed_order", getDefaultData.RouteMode)
	require.False(t, getDefaultData.AttemptTimeoutEnabled)
	require.Equal(t, 0, getDefaultData.AttemptTimeoutMinutes)
	require.False(t, getDefaultData.AllowLoss)
	require.Nil(t, getDefaultData.MaxLossAmount)
	require.Equal(t, 0, getDefaultData.BoundChannelCount)
	require.Equal(t, "", getDefaultData.PrimaryChannelName)
	require.Equal(t, "", getDefaultData.MinChannelCost)
	require.False(t, getDefaultData.ChannelAutoPriceStatus)

	invalidTimeoutRaw := h.rawRequest(http.MethodPatch, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-config", map[string]any{
		"smart_replenish_enabled": true,
		"attempt_timeout_enabled": true,
		"attempt_timeout_minutes": 0,
		"route_mode":              "fixed_order",
		"sync_cost_enabled":       true,
		"sync_goods_name_enabled": false,
		"allow_loss":              false,
		"max_loss_amount":         "",
		"is_bundle":               false,
	}, token)
	require.Equal(t, http.StatusOK, invalidTimeoutRaw.status)
	invalidTimeout := invalidTimeoutRaw.env
	require.Equal(t, 400, invalidTimeout.Code)

	updateResRaw := h.rawRequest(http.MethodPatch, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-config", map[string]any{
		"smart_replenish_enabled": true,
		"attempt_timeout_enabled": true,
		"attempt_timeout_minutes": 10,
		"route_mode":              "lowest_cost_first",
		"sync_cost_enabled":       true,
		"sync_goods_name_enabled": false,
		"allow_loss":              false,
		"max_loss_amount":         "2.0000",
		"is_bundle":               false,
	}, token)
	require.Equal(t, http.StatusOK, updateResRaw.status)
	updateRes := updateResRaw.env
	require.Equal(t, 0, updateRes.Code)

	// 插入两条绑定，验证 lowest_cost_first 的主绑定选择与 min_channel_cost 计算。
	accountAResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道A', 'test', 'test', 1, 1, 'https://example.com', 'token-a', 'secret', ?, ?)
`, now, now)
	require.NoError(t, err)
	accountAID, err := accountAResult.LastInsertId()
	require.NoError(t, err)

	accountBResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道B', 'test', 'test', 1, 1, 'https://example.com', 'token-b', 'secret', ?, ?)
`, now, now)
	require.NoError(t, err)
	accountBID, err := accountBResult.LastInsertId()
	require.NoError(t, err)

	_, err = h.app.Core().DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (goods_id, platform_account_id, supplier_goods_no, cost_price, dock_status, sort, created_at, updated_at)
VALUES (?, ?, 'A-001', '100.0000', 'enabled', 0, ?, ?),
       (?, ?, 'B-001', '27.1000',  'enabled', 0, ?, ?)
	`, goodsID, accountAID, now, now, goodsID, accountBID, now, now)
	require.NoError(t, err)

	getAfterUpdateRaw := h.rawRequest(http.MethodGet, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-config", nil, token)
	require.Equal(t, http.StatusOK, getAfterUpdateRaw.status)
	getAfterUpdate := getAfterUpdateRaw.env
	require.Equal(t, 0, getAfterUpdate.Code)
	var getAfterUpdateData struct {
		RouteMode          string  `json:"route_mode"`
		AllowLoss          bool    `json:"allow_loss"`
		MaxLossAmount      *string `json:"max_loss_amount"`
		BoundChannelCount  int     `json:"bound_channel_count"`
		PrimaryChannelName string  `json:"primary_channel_name"`
		MinChannelCost     string  `json:"min_channel_cost"`
	}
	require.NoError(t, json.Unmarshal(getAfterUpdate.Data, &getAfterUpdateData))
	require.Equal(t, "lowest_cost_first", getAfterUpdateData.RouteMode)
	require.False(t, getAfterUpdateData.AllowLoss)
	require.Nil(t, getAfterUpdateData.MaxLossAmount)
	require.Equal(t, 2, getAfterUpdateData.BoundChannelCount)
	require.Equal(t, "渠道B", getAfterUpdateData.PrimaryChannelName)
	require.Equal(t, "27.1000", getAfterUpdateData.MinChannelCost)

	listRaw := h.rawRequest(http.MethodGet, "/api/admin/products?page=1&page_size=20", nil, token)
	require.Equal(t, http.StatusOK, listRaw.status)
	require.Equal(t, 0, listRaw.env.Code)
	var listData struct {
		List []struct {
			ID                     int64    `json:"id"`
			BoundChannels          []string `json:"bound_channels"`
			BoundChannelCount      int      `json:"bound_channel_count"`
			PrimaryChannelName     string   `json:"primary_channel_name"`
			MinChannelCost         string   `json:"min_channel_cost"`
			ChannelAutoPriceStatus bool     `json:"channel_auto_price_status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRaw.env.Data, &listData))
	require.NotEmpty(t, listData.List)
	matchedIndex := -1
	for i := range listData.List {
		if listData.List[i].ID == goodsID {
			matchedIndex = i
			break
		}
	}
	require.NotEqual(t, -1, matchedIndex)
	require.Equal(t, []string{"渠道A", "渠道B"}, listData.List[matchedIndex].BoundChannels)
	require.Equal(t, 2, listData.List[matchedIndex].BoundChannelCount)
	require.Equal(t, "渠道B", listData.List[matchedIndex].PrimaryChannelName)
	require.Equal(t, "27.1000", listData.List[matchedIndex].MinChannelCost)
	require.False(t, listData.List[matchedIndex].ChannelAutoPriceStatus)
}
