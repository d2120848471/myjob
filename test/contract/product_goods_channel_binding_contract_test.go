package contract_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAPI_ProductGoodsChannelBindingPathsExposed(t *testing.T) {
	h := newTestHarness(t)

	res := h.rawRequest(http.MethodGet, "/api.json", nil, "")
	require.Equal(t, http.StatusOK, res.status)
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings/{binding_id}")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings:batch-status")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings:batch-delete")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings:reorder")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings/{binding_id}/auto-price")
	require.Contains(t, res.body, "/api/admin/product-goods/{goods_id}/channel-bindings:auto-price-batch")
}

func TestProductGoodsChannelBinding_CreateAndList_CalcCostPriceAndDisplayName(t *testing.T) {
	h := newTestHarness(t)
	token := h.loginAdmin(t)

	ctx := context.Background()
	now := time.Now()

	subjectResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('渠道绑定主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, err := subjectResult.LastInsertId()
	require.NoError(t, err)

	goodsCode := "T-CHBIND-001"
	goodsResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO product_goods (goods_code, brand_id, name, goods_type, supply_type, has_tax, subject_id, created_at, updated_at)
VALUES (?, 1, '渠道绑定测试商品', 'card_secret', 'channel', 1, ?, ?, ?)
`, goodsCode, subjectID, now, now)
	require.NoError(t, err)
	goodsID, err := goodsResult.LastInsertId()
	require.NoError(t, err)

	accountResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'test', '测试平台', 1, ?, 0, 'https://example.com', 'token-bind-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)
	accountID, err := accountResult.LastInsertId()
	require.NoError(t, err)

	// 配置税点：商品含税，渠道不含税 -> 需要 untaxed_to_taxed_rate。
	_, err = h.app.Core().DB().Exec(ctx, `
UPDATE system_config
SET config_value = '13'
WHERE config_key = 'trade.tax.untaxed_to_taxed_rate'
`)
	require.NoError(t, err)

	createRaw := h.rawRequest(http.MethodPost, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-bindings", map[string]any{
		"platform_account_id":  accountID,
		"supplier_goods_no":    "G001",
		"supplier_goods_name":  "上游商品A",
		"source_cost_price":    "100.0000",
		"dock_status":          "enabled",
		"sort":                 0,
		"weight":               0,
		"start_time":           "",
		"end_time":             "",
		"validate_template_id": nil,
	}, token)
	require.Equal(t, http.StatusOK, createRaw.status)
	require.Equal(t, 0, createRaw.env.Code)

	listRaw := h.rawRequest(http.MethodGet, "/api/admin/product-goods/"+int64ToString(goodsID)+"/channel-bindings", nil, token)
	require.Equal(t, http.StatusOK, listRaw.status)
	require.Equal(t, 0, listRaw.env.Code)

	var listData struct {
		List []struct {
			ID                  int64  `json:"id"`
			DisplayName         string `json:"display_name"`
			DockStatus          string `json:"dock_status"`
			PlatformAccountID   int64  `json:"platform_account_id"`
			PlatformAccountName string `json:"platform_account_name"`
			ProviderCode        string `json:"provider_code"`
			ProviderName        string `json:"provider_name"`
			SupplierGoodsNo     string `json:"supplier_goods_no"`
			SupplierGoodsName   string `json:"supplier_goods_name"`
			SourceCostPrice     string `json:"source_cost_price"`
			CostPrice           string `json:"cost_price"`
			TaxAdjustDirection  string `json:"tax_adjust_direction"`
			TaxAdjustRate       string `json:"tax_adjust_rate"`
			TaxAdjustAmount     string `json:"tax_adjust_amount"`
		} `json:"list"`
	}
	require.NoError(t, json.Unmarshal(listRaw.env.Data, &listData))
	require.Len(t, listData.List, 1)
	require.Equal(t, accountID, listData.List[0].PlatformAccountID)
	require.Equal(t, "渠道账号A", listData.List[0].PlatformAccountName)
	require.Equal(t, "test", listData.List[0].ProviderCode)
	require.Equal(t, "测试平台", listData.List[0].ProviderName)
	require.Equal(t, "G001", listData.List[0].SupplierGoodsNo)
	require.Equal(t, "上游商品A", listData.List[0].SupplierGoodsName)
	require.Equal(t, "enabled", listData.List[0].DockStatus)
	require.Equal(t, "100.0000", listData.List[0].SourceCostPrice)
	require.Equal(t, "113.0000", listData.List[0].CostPrice)
	require.Equal(t, "untaxed_to_taxed", listData.List[0].TaxAdjustDirection)
	require.Equal(t, "13.0000", listData.List[0].TaxAdjustRate)
	require.Equal(t, "13.0000", listData.List[0].TaxAdjustAmount)
	require.Equal(t, "上游商品A / 渠道绑定主体A / 测试平台", listData.List[0].DisplayName)
}
