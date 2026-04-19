package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_ProductGoodsInventoryConfigPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/products/{goodsId}/inventory-config")
}

func TestProductGoodsInventoryConfig_ReadSaveAndSummary(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "库存配置商品", "视频充值", "会员季卡")
	platformSubjectID := h.createSubject(t, token, "库存配置主体", 0)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "库存配置商品A", "29.9000")
	firstPlatformID := h.createSupplierPlatformAccount(t, token, "库存账号A", platformSubjectID, 0, "inventory-account-a")
	secondPlatformID := h.createSupplierPlatformAccount(t, token, "库存账号B", platformSubjectID, 0, "inventory-account-b")

	createFirst := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": firstPlatformID,
		"supplier_goods_no":   "INV-A",
		"supplier_goods_name": "库存商品A",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
		"order_weight":        "60.0000",
		"order_time_start":    "09:00",
		"order_time_end":      "18:00",
	}, token)
	require.Equal(t, 0, createFirst.Code)

	createSecond := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": secondPlatformID,
		"supplier_goods_no":   "INV-B",
		"supplier_goods_name": "库存商品B",
		"source_cost_price":   "12.0000",
		"dock_status":         1,
		"sort":                20,
		"order_weight":        "40.0000",
		"order_time_start":    "18:00",
		"order_time_end":      "02:00",
	}, token)
	require.Equal(t, 0, createSecond.Code)

	defaultRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", token)
	require.Equal(t, 0, defaultRes.Code)

	var defaultData struct {
		Config struct {
			SmartReorderEnabled   int    `json:"smart_reorder_enabled"`
			OrderStrategy         string `json:"order_strategy"`
			MaxLossAmount         string `json:"max_loss_amount"`
			ComboGoodsEnabled     int    `json:"combo_goods_enabled"`
			SyncGoodsNameEnabled  int    `json:"sync_goods_name_enabled"`
			ReorderTimeoutMinutes int    `json:"reorder_timeout_minutes"`
		} `json:"config"`
		OrderStrategyOptions []struct {
			Value string `json:"value"`
		} `json:"order_strategy_options"`
	}
	require.NoError(t, json.Unmarshal(defaultRes.Data, &defaultData))
	require.Equal(t, 0, defaultData.Config.SmartReorderEnabled)
	require.Equal(t, "fixed_order", defaultData.Config.OrderStrategy)
	require.Equal(t, "0.0000", defaultData.Config.MaxLossAmount)
	require.Equal(t, 0, defaultData.Config.ComboGoodsEnabled)
	require.Equal(t, 0, defaultData.Config.SyncGoodsNameEnabled)
	require.Equal(t, 0, defaultData.Config.ReorderTimeoutMinutes)
	require.Len(t, defaultData.OrderStrategyOptions, 5)

	saveRes := h.putJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"reorder_timeout_enabled": 1,
		"reorder_timeout_minutes": 30,
		"order_strategy":          "weighted_percent",
		"sync_cost_price_enabled": 1,
		"sync_goods_name_enabled": 1,
		"allow_loss_sale_enabled": 1,
		"max_loss_amount":         "2.5000",
		"combo_goods_enabled":     1,
	}, token)
	require.Equal(t, 0, saveRes.Code)

	repeatSaveRes := h.putJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"reorder_timeout_enabled": 1,
		"reorder_timeout_minutes": 30,
		"order_strategy":          "weighted_percent",
		"sync_cost_price_enabled": 1,
		"sync_goods_name_enabled": 1,
		"allow_loss_sale_enabled": 1,
		"max_loss_amount":         "2.5000",
		"combo_goods_enabled":     1,
	}, token)
	require.Equal(t, 0, repeatSaveRes.Code)

	updateSaveRes := h.putJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"reorder_timeout_enabled": 0,
		"reorder_timeout_minutes": 120,
		"order_strategy":          "fixed_order",
		"sync_cost_price_enabled": 0,
		"sync_goods_name_enabled": 0,
		"allow_loss_sale_enabled": 0,
		"max_loss_amount":         "9.9000",
		"combo_goods_enabled":     0,
	}, token)
	require.Equal(t, 0, updateSaveRes.Code)

	savedRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/inventory-config", token)
	require.Equal(t, 0, savedRes.Code)

	var savedData struct {
		Config struct {
			SmartReorderEnabled   int    `json:"smart_reorder_enabled"`
			ReorderTimeoutEnabled int    `json:"reorder_timeout_enabled"`
			ReorderTimeoutMinutes int    `json:"reorder_timeout_minutes"`
			OrderStrategy         string `json:"order_strategy"`
			SyncCostPriceEnabled  int    `json:"sync_cost_price_enabled"`
			SyncGoodsNameEnabled  int    `json:"sync_goods_name_enabled"`
			AllowLossSaleEnabled  int    `json:"allow_loss_sale_enabled"`
			MaxLossAmount         string `json:"max_loss_amount"`
			ComboGoodsEnabled     int    `json:"combo_goods_enabled"`
		} `json:"config"`
	}
	require.NoError(t, json.Unmarshal(savedRes.Data, &savedData))
	require.Equal(t, 1, savedData.Config.SmartReorderEnabled)
	require.Equal(t, 0, savedData.Config.ReorderTimeoutEnabled)
	require.Equal(t, 0, savedData.Config.ReorderTimeoutMinutes)
	require.Equal(t, "fixed_order", savedData.Config.OrderStrategy)
	require.Equal(t, 0, savedData.Config.SyncCostPriceEnabled)
	require.Equal(t, 0, savedData.Config.SyncGoodsNameEnabled)
	require.Equal(t, 0, savedData.Config.AllowLossSaleEnabled)
	require.Equal(t, "0.0000", savedData.Config.MaxLossAmount)
	require.Equal(t, 0, savedData.Config.ComboGoodsEnabled)

	bindingsRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, bindingsRes.Code)

	var bindingsData struct {
		Goods struct {
			InventoryConfigSummary struct {
				SmartReorderEnabled  int    `json:"smart_reorder_enabled"`
				OrderStrategy        string `json:"order_strategy"`
				SyncCostPriceEnabled int    `json:"sync_cost_price_enabled"`
				SyncGoodsNameEnabled int    `json:"sync_goods_name_enabled"`
				AllowLossSaleEnabled int    `json:"allow_loss_sale_enabled"`
				ComboGoodsEnabled    int    `json:"combo_goods_enabled"`
			} `json:"inventory_config_summary"`
		} `json:"goods"`
	}
	require.NoError(t, json.Unmarshal(bindingsRes.Data, &bindingsData))
	require.Equal(t, 1, bindingsData.Goods.InventoryConfigSummary.SmartReorderEnabled)
	require.Equal(t, "fixed_order", bindingsData.Goods.InventoryConfigSummary.OrderStrategy)
	require.Equal(t, 0, bindingsData.Goods.InventoryConfigSummary.SyncCostPriceEnabled)
	require.Equal(t, 0, bindingsData.Goods.InventoryConfigSummary.SyncGoodsNameEnabled)
	require.Equal(t, 0, bindingsData.Goods.InventoryConfigSummary.AllowLossSaleEnabled)
	require.Equal(t, 0, bindingsData.Goods.InventoryConfigSummary.ComboGoodsEnabled)

	logsRes := h.getJSON("/api/admin/logs/operations?page=1&page_size=20&keyword=库存配置", token)
	require.Equal(t, 0, logsRes.Code)

	var logsData struct {
		List []struct {
			Description string `json:"description"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(logsRes.Data, &logsData))
	require.NotEmpty(t, logsData.List)
	require.Contains(t, logsData.List[0].Description, "修改商品库存配置")
}

func TestProductGoodsInventoryConfig_ValidationAndPermission(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "库存配置校验商品", "视频充值", "会员月卡")
	platformSubjectID := h.createSubject(t, token, "库存配置校验主体", 0)
	channelGoodsID := h.createChannelProductGoods(t, token, leafBrandID, "库存配置校验商品A", "19.9000")
	platformID := h.createSupplierPlatformAccount(t, token, "库存校验账号", platformSubjectID, 0, "inventory-validate-account")

	createBinding := h.postJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "INV-VALIDATE",
		"supplier_goods_name": "库存校验商品",
		"source_cost_price":   "8.0000",
		"dock_status":         1,
		"sort":                10,
		"order_weight":        "80.0000",
		"order_time_start":    "22:00",
		"order_time_end":      "02:00",
	}, token)
	require.Equal(t, 0, createBinding.Code)

	invalidFlag := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 2,
		"order_strategy":        "fixed_order",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 400, invalidFlag.Code)

	invalidStrategy := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "legacy_random",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 400, invalidStrategy.Code)

	invalidTimeout := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"reorder_timeout_enabled": 1,
		"reorder_timeout_minutes": 1441,
		"order_strategy":          "fixed_order",
		"max_loss_amount":         "0.0000",
	}, token)
	require.Equal(t, 400, invalidTimeout.Code)

	invalidAmount := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled":   1,
		"order_strategy":          "fixed_order",
		"allow_loss_sale_enabled": 1,
		"max_loss_amount":         "-1",
	}, token)
	require.Equal(t, 400, invalidAmount.Code)

	invalidWeightStrategy := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "weighted_percent",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 400, invalidWeightStrategy.Code)

	now := h.app.Core().Now()
	_, err := h.app.Core().DB().GetCore().Exec(context.Background(), `
UPDATE product_goods_channel_binding
SET order_weight = ?, updated_at = ?
WHERE goods_id = ? AND platform_account_id = ? AND is_deleted = 0
`, "100.0000", now, channelGoodsID, platformID)
	require.NoError(t, err)

	weightedPercentRes := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "weighted_percent",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 0, weightedPercentRes.Code)

	timeWindowRes := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "time_window",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 0, timeWindowRes.Code)

	_, err = h.app.Core().DB().GetCore().Exec(context.Background(), `
UPDATE product_goods_channel_binding
SET order_time_end = NULL, updated_at = ?
WHERE goods_id = ? AND platform_account_id = ? AND is_deleted = 0
`, now, channelGoodsID, platformID)
	require.NoError(t, err)

	timeWindowMissingEnd := h.putJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "time_window",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 400, timeWindowMissingEnd.Code)

	nonChannelID := h.createChannelProductGoods(t, token, leafBrandID, "非渠道商品", "9.9000")
	_, err = h.app.Core().DB().GetCore().Exec(context.Background(), `
UPDATE product_goods
SET supply_type = ?, updated_at = ?
WHERE id = ?
`, "manual", h.app.Core().Now(), nonChannelID)
	require.NoError(t, err)

	nonChannelSave := h.putJSON("/api/admin/products/"+int64ToString(nonChannelID)+"/inventory-config", map[string]any{
		"smart_reorder_enabled": 1,
		"order_strategy":        "fixed_order",
		"max_loss_amount":       "0.0000",
	}, token)
	require.Equal(t, 400, nonChannelSave.Code)

	limitedToken := h.createLimitedUserToken(t, token, 0)
	forbiddenRes := h.getJSON("/api/admin/products/"+int64ToString(channelGoodsID)+"/inventory-config", limitedToken)
	require.Equal(t, 403, forbiddenRes.Code)
}
