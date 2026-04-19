package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type channelBindingContractRow struct {
	ID                 int64  `json:"id"`
	PlatformAccountID  int64  `json:"platform_account_id"`
	SupplierGoodsName  string `json:"supplier_goods_name"`
	SourceCostPrice    string `json:"source_cost_price"`
	CostPrice          string `json:"cost_price"`
	EffectiveSellPrice string `json:"effective_sell_price"`
	DockStatus         int    `json:"dock_status"`
	Sort               int    `json:"sort"`
	OrderWeight        string `json:"order_weight"`
	OrderTimeStart     string `json:"order_time_start"`
	OrderTimeEnd       string `json:"order_time_end"`
	IsAutoChange       int    `json:"is_auto_change"`
	AddType            string `json:"add_type"`
	DefaultPrice       string `json:"default_price"`
}

func TestOpenAPI_ProductGoodsChannelPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/products/{goodsId}/channel-bindings")
	require.Contains(t, res.body, "/api/admin/products/{goodsId}/channel-bindings/form-options")
	require.Contains(t, res.body, "/api/admin/products/{goodsId}/channel-bindings/{bindingId}")
	require.Contains(t, res.body, "/api/admin/products/{goodsId}/channel-bindings/{bindingId}/auto-price")
	require.Contains(t, res.body, "order_weight")
	require.Contains(t, res.body, "order_time_start")
	require.Contains(t, res.body, "order_time_end")
}

func TestProductGoodsChannelBindingFlows(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品", "视频充值", "会员周卡")
	productTemplateID := h.createProductTemplate(t, token, "渠道模板")
	platformSubjectID := h.createSubject(t, token, "渠道主体", 1)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品A", "18.8000")

	untaxedPlatformID := h.createSupplierPlatformAccount(t, token, "渠道未税账号", platformSubjectID, 0, "untaxed-account")
	taxedPlatformID := h.createSupplierPlatformAccount(t, token, "渠道含税账号", platformSubjectID, 1, "taxed-account")

	formOptionsBefore := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/form-options", token)
	require.Equal(t, 0, formOptionsBefore.Code)

	var formOptionsBeforeData struct {
		PlatformAccounts []struct {
			ID int64 `json:"id"`
		} `json:"platform_accounts"`
		ValidateTemplates []struct {
			ID int64 `json:"id"`
		} `json:"validate_templates"`
	}
	require.NoError(t, json.Unmarshal(formOptionsBefore.Data, &formOptionsBeforeData))
	require.Len(t, formOptionsBeforeData.PlatformAccounts, 2)
	require.Contains(t, formOptionsBeforeData.ValidateTemplates, struct {
		ID int64 `json:"id"`
	}{ID: productTemplateID})

	missingRateRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id":  taxedPlatformID,
		"supplier_goods_no":    "SKU-TAXED-1",
		"supplier_goods_name":  "含税渠道商品",
		"source_cost_price":    "10.0000",
		"validate_template_id": productTemplateID,
		"dock_status":          1,
		"sort":                 20,
	}, token)
	require.Equal(t, 400, missingRateRes.Code)

	h.saveFinanceTaxConfig(t, token, "4.5", "3.8")

	createFirst := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id":  untaxedPlatformID,
		"supplier_goods_no":    "SKU-UNTAXED-1",
		"supplier_goods_name":  "未税渠道商品",
		"source_cost_price":    "10.0000",
		"validate_template_id": productTemplateID,
		"dock_status":          1,
		"sort":                 10,
		"order_weight":         "60.0000",
		"order_time_start":     "09:00",
		"order_time_end":       "18:00",
	}, token)
	require.Equal(t, 0, createFirst.Code)

	createDuplicate := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": untaxedPlatformID,
		"supplier_goods_no":   "SKU-UNTAXED-1",
		"supplier_goods_name": "重复绑定",
		"source_cost_price":   "11.0000",
		"dock_status":         1,
		"sort":                11,
	}, token)
	require.Equal(t, 400, createDuplicate.Code)

	createSecond := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id":  taxedPlatformID,
		"supplier_goods_no":    "SKU-TAXED-1",
		"supplier_goods_name":  "含税渠道商品",
		"source_cost_price":    "10.0000",
		"validate_template_id": productTemplateID,
		"dock_status":          1,
		"sort":                 20,
		"order_weight":         "40.0000",
		"order_time_start":     "18:00",
		"order_time_end":       "02:00",
	}, token)
	require.Equal(t, 0, createSecond.Code)

	var createFirstData struct {
		ID int64 `json:"id"`
	}
	var createSecondData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createFirst.Data, &createFirstData))
	require.NoError(t, json.Unmarshal(createSecond.Data, &createSecondData))

	bindingsRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, bindingsRes.Code)

	var bindingsData struct {
		Goods struct {
			ID               int64  `json:"id"`
			Name             string `json:"name"`
			HasTax           int    `json:"has_tax"`
			DefaultSellPrice string `json:"default_sell_price"`
		} `json:"goods"`
		List []channelBindingContractRow `json:"list"`
	}
	require.NoError(t, json.Unmarshal(bindingsRes.Data, &bindingsData))
	require.Equal(t, goodsID, bindingsData.Goods.ID)
	require.Equal(t, "渠道商品A", bindingsData.Goods.Name)
	require.Equal(t, 0, bindingsData.Goods.HasTax)
	require.Equal(t, "18.8000", bindingsData.Goods.DefaultSellPrice)
	require.Len(t, bindingsData.List, 2)

	firstBinding := findBindingByID(t, bindingsData.List, createFirstData.ID)
	require.Equal(t, "10.0000", firstBinding.SourceCostPrice)
	require.Equal(t, "10.0000", firstBinding.CostPrice)
	require.Equal(t, "18.8000", firstBinding.EffectiveSellPrice)
	require.Equal(t, "60.0000", firstBinding.OrderWeight)
	require.Equal(t, "09:00", firstBinding.OrderTimeStart)
	require.Equal(t, "18:00", firstBinding.OrderTimeEnd)
	require.Equal(t, 0, firstBinding.IsAutoChange)

	secondBinding := findBindingByID(t, bindingsData.List, createSecondData.ID)
	require.Equal(t, "10.0000", secondBinding.SourceCostPrice)
	require.Equal(t, "9.6200", secondBinding.CostPrice)
	require.Equal(t, "18.8000", secondBinding.EffectiveSellPrice)
	require.Equal(t, "40.0000", secondBinding.OrderWeight)
	require.Equal(t, "18:00", secondBinding.OrderTimeStart)
	require.Equal(t, "02:00", secondBinding.OrderTimeEnd)
	require.Equal(t, 0, secondBinding.IsAutoChange)

	listRes := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=渠道商品A", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			ID                           int64    `json:"id"`
			BoundChannels                []string `json:"bound_channels"`
			BoundChannelCount            int      `json:"bound_channel_count"`
			PrimaryChannelName           string   `json:"primary_channel_name"`
			MinChannelCost               string   `json:"min_channel_cost"`
			MinChannelEffectiveSellPrice string   `json:"min_channel_effective_sell_price"`
			ChannelAutoPriceStatus       int      `json:"channel_auto_price_status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, goodsID, listData.List[0].ID)
	require.Equal(t, []string{"渠道未税账号", "渠道含税账号"}, listData.List[0].BoundChannels)
	require.Equal(t, 2, listData.List[0].BoundChannelCount)
	require.Equal(t, "渠道未税账号", listData.List[0].PrimaryChannelName)
	require.Equal(t, "9.6200", listData.List[0].MinChannelCost)
	require.Equal(t, "18.8000", listData.List[0].MinChannelEffectiveSellPrice)
	require.Equal(t, 0, listData.List[0].ChannelAutoPriceStatus)

	autoPriceRes := h.patchJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(createSecondData.ID)+"/auto-price", map[string]any{
		"is_auto_change": 1,
		"add_type":       "fixed",
		"default_price":  "1.3000",
	}, token)
	require.Equal(t, 0, autoPriceRes.Code)

	afterAutoPrice := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, afterAutoPrice.Code)
	require.NoError(t, json.Unmarshal(afterAutoPrice.Data, &bindingsData))
	secondBinding = findBindingByID(t, bindingsData.List, createSecondData.ID)
	require.Equal(t, 1, secondBinding.IsAutoChange)
	require.Equal(t, "fixed", secondBinding.AddType)
	require.Equal(t, "1.3000", secondBinding.DefaultPrice)
	require.Equal(t, "10.9200", secondBinding.EffectiveSellPrice)

	updateBinding := h.patchJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(createSecondData.ID), map[string]any{
		"platform_account_id":  taxedPlatformID,
		"supplier_goods_no":    "SKU-TAXED-1",
		"supplier_goods_name":  "含税渠道商品-编辑",
		"source_cost_price":    "11.0000",
		"validate_template_id": productTemplateID,
		"dock_status":          1,
		"sort":                 5,
		"order_weight":         "55.0000",
		"order_time_start":     "20:00",
		"order_time_end":       "03:00",
	}, token)
	require.Equal(t, 0, updateBinding.Code)

	afterUpdateBindings := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, afterUpdateBindings.Code)
	require.NoError(t, json.Unmarshal(afterUpdateBindings.Data, &bindingsData))
	secondBinding = findBindingByID(t, bindingsData.List, createSecondData.ID)
	require.Equal(t, "55.0000", secondBinding.OrderWeight)
	require.Equal(t, "20:00", secondBinding.OrderTimeStart)
	require.Equal(t, "03:00", secondBinding.OrderTimeEnd)

	afterUpdateList := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=渠道商品A", token)
	require.Equal(t, 0, afterUpdateList.Code)
	require.NoError(t, json.Unmarshal(afterUpdateList.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, "渠道含税账号", listData.List[0].PrimaryChannelName)
	require.Equal(t, "10.0000", listData.List[0].MinChannelCost)
	require.Equal(t, "11.8820", listData.List[0].MinChannelEffectiveSellPrice)
	require.Equal(t, 1, listData.List[0].ChannelAutoPriceStatus)

	deleteBinding := h.deleteJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(createSecondData.ID), token)
	require.Equal(t, 0, deleteBinding.Code)

	afterDeleteList := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=渠道商品A", token)
	require.Equal(t, 0, afterDeleteList.Code)
	require.NoError(t, json.Unmarshal(afterDeleteList.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, []string{"渠道未税账号"}, listData.List[0].BoundChannels)
	require.Equal(t, 1, listData.List[0].BoundChannelCount)
	require.Equal(t, "渠道未税账号", listData.List[0].PrimaryChannelName)
	require.Equal(t, "10.0000", listData.List[0].MinChannelCost)
	require.Equal(t, "18.8000", listData.List[0].MinChannelEffectiveSellPrice)
	require.Equal(t, 0, listData.List[0].ChannelAutoPriceStatus)
}

func TestProductGoodsChannelBinding_AllowsCreateWhenTaxConfigNotNeededOrOnlyOneDirectionConfigured(t *testing.T) {
	t.Run("税态一致时不要求财务税率", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)

		_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品-税率一致", "视频充值", "会员周卡")
		platformSubjectID := h.createSubject(t, token, "渠道主体-未税", 0)
		goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品-未税", "18.8000")
		untaxedPlatformID := h.createSupplierPlatformAccount(t, token, "渠道未税账号-一致", platformSubjectID, 0, "untaxed-same-tax")

		createRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
			"platform_account_id": untaxedPlatformID,
			"supplier_goods_no":   "SKU-UNTAXED-SAME",
			"supplier_goods_name": "未税渠道商品-一致",
			"source_cost_price":   "10.0000",
			"dock_status":         1,
			"sort":                1,
		}, token)
		require.Equal(t, 0, createRes.Code)
	})

	t.Run("只需要一个税率方向时允许缺少另一个方向", func(t *testing.T) {
		h := newTestHarness(t)
		token := h.loginAdmin(t)

		_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品-单向税率", "视频充值", "会员周卡")
		platformSubjectID := h.createSubject(t, token, "渠道主体-含税", 1)
		goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品-单向税率", "18.8000")
		taxedPlatformID := h.createSupplierPlatformAccount(t, token, "渠道含税账号-单向税率", platformSubjectID, 1, "taxed-single-rate")

		h.saveFinanceTaxConfigDirect(t, map[string]string{
			"finance_tax_inclusive_rate": "3.8",
		})

		createRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
			"platform_account_id": taxedPlatformID,
			"supplier_goods_no":   "SKU-TAXED-SINGLE-RATE",
			"supplier_goods_name": "含税渠道商品-单向税率",
			"source_cost_price":   "10.0000",
			"dock_status":         1,
			"sort":                1,
		}, token)
		require.Equal(t, 0, createRes.Code)
	})
}

func TestProductGoodsChannelBinding_HidesSoftDeletedPlatformAccounts(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品-软删平台", "视频充值", "会员周卡")
	platformSubjectID := h.createSubject(t, token, "渠道主体-软删平台", 0)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品-软删平台", "18.8000")
	platformID := h.createSupplierPlatformAccount(t, token, "渠道账号-软删平台", platformSubjectID, 0, "soft-delete-platform")

	createRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "SKU-SOFT-DELETE-PLATFORM",
		"supplier_goods_name": "软删平台渠道商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 0, createRes.Code)

	h.softDeleteSupplierPlatformAccount(t, platformID)

	bindingsRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, bindingsRes.Code)

	var bindingsData struct {
		List []channelBindingContractRow `json:"list"`
	}
	require.NoError(t, json.Unmarshal(bindingsRes.Data, &bindingsData))
	require.Empty(t, bindingsData.List)

	listRes := h.getJSON("/api/admin/products?page=1&page_size=20&keyword=渠道商品-软删平台", token)
	require.Equal(t, 0, listRes.Code)

	var listData struct {
		List []struct {
			BoundChannels                []string `json:"bound_channels"`
			BoundChannelCount            int      `json:"bound_channel_count"`
			PrimaryChannelName           string   `json:"primary_channel_name"`
			MinChannelCost               string   `json:"min_channel_cost"`
			MinChannelEffectiveSellPrice string   `json:"min_channel_effective_sell_price"`
			ChannelAutoPriceStatus       int      `json:"channel_auto_price_status"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRes.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Empty(t, listData.List[0].BoundChannels)
	require.Equal(t, 0, listData.List[0].BoundChannelCount)
	require.Empty(t, listData.List[0].PrimaryChannelName)
	require.Empty(t, listData.List[0].MinChannelCost)
	require.Empty(t, listData.List[0].MinChannelEffectiveSellPrice)
	require.Equal(t, 0, listData.List[0].ChannelAutoPriceStatus)
}

func TestProductGoodsChannelBinding_RejectsDisabledPlatformAccounts(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品-停用平台", "视频充值", "会员周卡")
	platformSubjectID := h.createSubject(t, token, "渠道主体-停用平台", 0)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品-停用平台", "18.8000")
	disabledPlatformID := h.createSupplierPlatformAccount(t, token, "渠道账号-停用", platformSubjectID, 0, "disabled-platform")
	enabledPlatformID := h.createSupplierPlatformAccount(t, token, "渠道账号-启用", platformSubjectID, 0, "enabled-platform")

	createBindingRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": disabledPlatformID,
		"supplier_goods_no":   "SKU-DISABLED-EXISTING",
		"supplier_goods_name": "停用平台既有绑定",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 0, createBindingRes.Code)

	var createBindingData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createBindingRes.Data, &createBindingData))
	require.NotZero(t, createBindingData.ID)

	h.updateSupplierPlatformStatus(t, token, disabledPlatformID, "渠道账号-停用", platformSubjectID, 0, "disabled-platform", 0)

	formOptionsRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/form-options", token)
	require.Equal(t, 0, formOptionsRes.Code)

	var formOptionsData struct {
		PlatformAccounts []struct {
			ID int64 `json:"id"`
		} `json:"platform_accounts"`
	}
	require.NoError(t, json.Unmarshal(formOptionsRes.Data, &formOptionsData))
	require.Contains(t, formOptionsData.PlatformAccounts, struct {
		ID int64 `json:"id"`
	}{ID: enabledPlatformID})
	require.NotContains(t, formOptionsData.PlatformAccounts, struct {
		ID int64 `json:"id"`
	}{ID: disabledPlatformID})

	bindingListRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, bindingListRes.Code)

	var bindingListData struct {
		List []channelBindingContractRow `json:"list"`
	}
	require.NoError(t, json.Unmarshal(bindingListRes.Data, &bindingListData))
	require.Empty(t, bindingListData.List)

	createDisabledRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": disabledPlatformID,
		"supplier_goods_no":   "SKU-DISABLED-PLATFORM",
		"supplier_goods_name": "停用平台渠道商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 400, createDisabledRes.Code)

	createEnabledRes := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": enabledPlatformID,
		"supplier_goods_no":   "SKU-ENABLED-PLATFORM",
		"supplier_goods_name": "启用平台渠道商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 0, createEnabledRes.Code)

	var createEnabledData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createEnabledRes.Data, &createEnabledData))
	require.NotZero(t, createEnabledData.ID)

	updateToDisabledRes := h.patchJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(createEnabledData.ID), map[string]any{
		"platform_account_id": disabledPlatformID,
		"supplier_goods_no":   "SKU-ENABLED-PLATFORM",
		"supplier_goods_name": "启用平台渠道商品",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                1,
	}, token)
	require.Equal(t, 400, updateToDisabledRes.Code)
}

func TestProductGoodsChannelBinding_UpdatePreservesExplicitZeroSort(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	_, _, leafBrandID := h.createBrandPath(t, token, "渠道商品-sort", "视频充值", "会员周卡")
	platformSubjectID := h.createSubject(t, token, "渠道主体-sort", 0)
	goodsID := h.createChannelProductGoods(t, token, leafBrandID, "渠道商品-sort", "18.8000")
	platformID := h.createSupplierPlatformAccount(t, token, "渠道账号-sort", platformSubjectID, 0, "sort-platform")

	createFirst := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "SKU-SORT-1",
		"supplier_goods_name": "渠道商品-sort-1",
		"source_cost_price":   "10.0000",
		"dock_status":         1,
		"sort":                10,
	}, token)
	require.Equal(t, 0, createFirst.Code)

	createSecond := h.postJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "SKU-SORT-2",
		"supplier_goods_name": "渠道商品-sort-2",
		"source_cost_price":   "11.0000",
		"dock_status":         1,
		"sort":                20,
	}, token)
	require.Equal(t, 0, createSecond.Code)

	var createSecondData struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(createSecond.Data, &createSecondData))

	updateRes := h.patchJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings/"+int64ToString(createSecondData.ID), map[string]any{
		"platform_account_id": platformID,
		"supplier_goods_no":   "SKU-SORT-2",
		"supplier_goods_name": "渠道商品-sort-2",
		"source_cost_price":   "11.0000",
		"dock_status":         1,
		"sort":                0,
	}, token)
	require.Equal(t, 0, updateRes.Code)

	bindingsRes := h.getJSON("/api/admin/products/"+int64ToString(goodsID)+"/channel-bindings", token)
	require.Equal(t, 0, bindingsRes.Code)

	var bindingsData struct {
		List []channelBindingContractRow `json:"list"`
	}
	require.NoError(t, json.Unmarshal(bindingsRes.Data, &bindingsData))
	require.Len(t, bindingsData.List, 2)
	require.Equal(t, createSecondData.ID, bindingsData.List[0].ID)
	require.Equal(t, 0, bindingsData.List[0].Sort)
}

func (h *testHarness) createChannelProductGoods(t *testing.T, token string, brandID int64, name, defaultSellPrice string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/products", map[string]any{
		"brand_id":           brandID,
		"name":               name,
		"goods_type":         "card_secret",
		"supply_type":        "channel",
		"is_export":          1,
		"is_douyin":          0,
		"has_tax":            0,
		"exception_notify":   1,
		"balance_limit":      "0",
		"default_sell_price": defaultSellPrice,
		"min_purchase_qty":   1,
		"max_purchase_qty":   1,
		"status":             1,
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) createSupplierPlatformAccount(t *testing.T, token, name string, subjectID int64, hasTax int, tokenID string) int64 {
	t.Helper()
	res := h.postJSON("/api/admin/supplier-platforms", map[string]any{
		"name":             name,
		"domain":           tokenID + ".test.local",
		"backup_domain":    tokenID + ".backup.local",
		"type_id":          35,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
	}, token)
	require.Equal(t, 0, res.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(res.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}

func (h *testHarness) updateSupplierPlatformStatus(t *testing.T, token string, id int64, name string, subjectID int64, hasTax int, tokenID string, status int) {
	t.Helper()
	res := h.putJSON("/api/admin/supplier-platforms/"+int64ToString(id), map[string]any{
		"name":             name,
		"domain":           tokenID + ".test.local",
		"backup_domain":    tokenID + ".backup.local",
		"type_id":          35,
		"subject_id":       subjectID,
		"has_tax":          hasTax,
		"token_id":         tokenID,
		"secret_key":       "secret-key",
		"threshold_amount": "5000.0000",
		"sort":             5,
		"crowd_name":       "运营群",
		"status":           status,
	}, token)
	require.Equal(t, 0, res.Code)
}

func (h *testHarness) saveFinanceTaxConfig(t *testing.T, token, exclusiveRate, inclusiveRate string) {
	t.Helper()
	res := h.rawRequest(http.MethodPut, "/api/admin/settings/system", map[string]any{
		"group": "finance",
		"items": []map[string]any{
			{"key": "tax_exclusive_rate", "value": exclusiveRate},
			{"key": "tax_inclusive_rate", "value": inclusiveRate},
		},
	}, token)
	require.Equal(t, http.StatusOK, res.status)
	require.Equal(t, 0, res.env.Code)
}

func (h *testHarness) saveFinanceTaxConfigDirect(t *testing.T, items map[string]string) {
	t.Helper()
	now := h.app.Core().Now()
	for key, value := range items {
		_, err := h.app.Core().DB().GetCore().Exec(context.Background(), `
INSERT INTO system_config (config_key, config_value, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), description = VALUES(description), updated_at = VALUES(updated_at)
`, key, value, key, now, now)
		require.NoError(t, err)
	}
}

func (h *testHarness) softDeleteSupplierPlatformAccount(t *testing.T, id int64) {
	t.Helper()
	now := h.app.Core().Now()
	_, err := h.app.Core().DB().GetCore().Exec(context.Background(), `
UPDATE supplier_platform_account
SET is_deleted = 1, deleted_at = ?, updated_at = ?
WHERE id = ?
`, now, now, id)
	require.NoError(t, err)
}

func findBindingByID(t *testing.T, list []channelBindingContractRow, id int64) channelBindingContractRow {
	t.Helper()
	for _, item := range list {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("binding %d not found", id)
	var zero channelBindingContractRow
	return zero
}
