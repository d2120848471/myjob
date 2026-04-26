package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

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
