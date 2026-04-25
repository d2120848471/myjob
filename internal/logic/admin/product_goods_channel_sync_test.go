package adminlogic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/model/entity"

	"github.com/stretchr/testify/require"
)

func TestSyncChannelBindingsOnceUpdatesNameAndTaxedCost(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		require.Equal(t, "/dockapiv3/goods/details", r.URL.Path)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		require.Equal(t, "merchant001", payload["userid"])
		require.Equal(t, "SKU-100", payload["goodsid"])
		_, _ = w.Write([]byte(`{"code":1,"message":"ok","data":{"goodsid":"SKU-100","goodsname":"上游新名字","goodsprice":"11.0000","goodsstatus":1}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 0)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 1, result.Scanned)
	require.Equal(t, 1, result.Updated)
	require.Equal(t, int64(1), requestCount.Load())

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "上游新名字", row.SupplierGoodsName)
	require.Equal(t, "11.0000", row.SourceCostPrice)
	require.Equal(t, "11.4950", row.CostPrice)
	require.Equal(t, taxAdjustDirectionUntaxedToTaxed, row.TaxAdjustDirection)
	require.Equal(t, "4.5000", row.TaxAdjustRate)
	require.Equal(t, "0.4950", row.TaxAdjustAmount)
	require.Equal(t, 1, row.IsAutoChange)
	require.Equal(t, "0.2000", row.DefaultPrice)
}

func TestSyncChannelBindingsOnceHonorsSwitchesAndKeepsManualValues(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"message":"ok","data":{"goodsid":"SKU-100","goodsname":"上游新名字","goodsprice":"12.0000"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 0, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 1, result.Updated)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "上游新名字", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
	require.Equal(t, "10.0000", row.CostPrice)
}

func TestSyncChannelBindingsOnceSkipsWhenSwitchesClosed(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"不应覆盖","goodsprice":"99"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 0, 0, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 0, result.Scanned)
	require.Equal(t, int64(0), requestCount.Load())

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "人工名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
}

func TestSyncChannelBindingsOnceSkipsDisabledGoods(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"不应覆盖","goodsprice":"99"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1, 0)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 0, result.Scanned)
	require.Equal(t, int64(0), requestCount.Load())

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "人工名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
}

func TestSyncChannelBindingsOnceDoesNotOverwriteWhenSwitchClosesDuringFetch(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	goodsID := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, updateErr := core.DB().Exec(ctx, `
UPDATE product_goods_channel_config
SET sync_cost_price_enabled = 0, sync_goods_name_enabled = 0, updated_at = ?
WHERE goods_id = ?
`, core.Now(), goodsID)
		require.NoError(t, updateErr)
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"不应覆盖","goodsprice":"99.0000"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID = seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 0, result.Updated)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "人工名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
	require.Equal(t, "10.0000", row.CostPrice)
}

func TestSyncChannelBindingsOnceSkipsMismatchedSupplierGoodsNo(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"OTHER-SKU","goodsname":"不应覆盖","goodsprice":"99.0000"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 0, result.Updated)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "人工名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
	require.Equal(t, "10.0000", row.CostPrice)
}

func TestSyncChannelBindingsOnceScansBeyondLimit(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	var requestCount atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		goodsNo := payload["goodsid"].(string)
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"` + goodsNo + `","goodsname":"上游-` + goodsNo + `","goodsprice":"11.0000"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	seedProductGoodsSyncExtraBinding(t, core, goodsID, "SKU-101")
	seedProductGoodsSyncExtraBinding(t, core, goodsID, "SKU-102")
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 1})
	require.NoError(t, err)
	require.Equal(t, 3, result.Scanned)
	require.Equal(t, 3, result.Updated)
	require.Equal(t, int64(3), requestCount.Load())

	value, err := core.DB().GetCore().GetValue(ctx, `
SELECT COUNT(*)
FROM product_goods_channel_binding
WHERE goods_id = ? AND supplier_goods_name LIKE '上游-SKU-%'
`, goodsID)
	require.NoError(t, err)
	require.Equal(t, 3, value.Int())
}

func TestSyncChannelBindingsOnceDoesNotOverwriteWithBadUpstreamData(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"","goodsprice":"abc"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 1, result.Failed)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "人工名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
	require.Equal(t, "10.0000", row.CostPrice)
}

func TestSyncChannelBindingsOnceUpdatesNameWhenPriceInvalid(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"只同步名称","goodsprice":"abc"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 0)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	result, err := logic.SyncChannelBindingsOnce(ctx, ProductGoodsChannelSyncOptions{GoodsID: goodsID, Limit: 200})
	require.NoError(t, err)
	require.Equal(t, 1, result.Updated)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "只同步名称", row.SupplierGoodsName)
	require.Equal(t, "10.0000", row.SourceCostPrice)
	require.Equal(t, "10.0000", row.CostPrice)
}

func TestSaveInventoryConfigTriggersImmediateSyncWhenSwitchEnabled(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	ctx := context.Background()
	seedProductGoodsSyncTaxConfig(t, core)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":1,"message":"ok","data":{"goodsid":"SKU-100","goodsname":"保存后新名字","goodsprice":"11.0000"}}`))
	}))
	t.Cleanup(server.Close)

	goodsID := seedProductGoodsSyncGoods(t, core, 0, 0, strings.TrimPrefix(server.URL, "http://"), 1, 0)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)

	_, err = logic.SaveInventoryConfig(ctx, &adminapi.ProductGoodsInventoryConfigSaveReq{
		GoodsId:               goodsID,
		SmartReorderEnabled:   0,
		ReorderTimeoutEnabled: 0,
		ReorderTimeoutMinutes: 0,
		OrderStrategy:         productGoodsOrderStrategyFixedOrder,
		SyncCostPriceEnabled:  1,
		SyncGoodsNameEnabled:  1,
		AllowLossSaleEnabled:  0,
		MaxLossAmount:         "0",
		ComboGoodsEnabled:     0,
	}, entity.AdminUser{ID: 1, Username: "admin"}, "127.0.0.1")
	require.NoError(t, err)

	row := loadProductGoodsSyncBinding(t, core, goodsID)
	require.Equal(t, "保存后新名字", row.SupplierGoodsName)
	require.Equal(t, "11.0000", row.SourceCostPrice)
	require.Equal(t, "11.4950", row.CostPrice)
}

func TestProductGoodsChannelSyncWorkerSkipsOverlappingRuns(t *testing.T) {
	worker := NewProductGoodsChannelSyncWorker(&ProductGoodsLogic{}, ProductGoodsChannelSyncWorkerOptions{
		Interval: time.Hour,
		Limit:    200,
	})
	require.True(t, worker.tryBeginRun())
	require.False(t, worker.tryBeginRun())
	worker.finishRun()
	require.True(t, worker.tryBeginRun())
	worker.finishRun()
}

func TestProductGoodsChannelSyncWorkerStopCancelsRunningRequest(t *testing.T) {
	core, err := app.NewTestCore()
	require.NoError(t, err)
	t.Cleanup(func() { _ = core.Close() })
	seedProductGoodsSyncTaxConfig(t, core)

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-requestStarted:
		default:
			close(requestStarted)
		}
		select {
		case <-r.Context().Done():
			return
		case <-releaseRequest:
			_, _ = w.Write([]byte(`{"code":1,"data":{"goodsid":"SKU-100","goodsname":"上游新名字","goodsprice":"11.0000"}}`))
		}
	}))
	t.Cleanup(server.Close)

	seedProductGoodsSyncGoods(t, core, 1, 1, strings.TrimPrefix(server.URL, "http://"), 1, 1)
	logic := newProductGoodsSyncTestLogic(t, core, server.URL)
	worker := NewProductGoodsChannelSyncWorker(logic, ProductGoodsChannelSyncWorkerOptions{
		Interval: time.Millisecond,
		Limit:    200,
	})
	worker.Start()
	<-requestStarted

	stopped := make(chan struct{})
	go func() {
		worker.Stop()
		close(stopped)
	}()

	stopReturned := false
	select {
	case <-stopped:
		stopReturned = true
	case <-time.After(200 * time.Millisecond):
	}
	close(releaseRequest)
	if !stopReturned {
		<-stopped
		t.Fatal("worker Stop 没有取消正在执行的上游请求")
	}
}

type productGoodsSyncRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (t productGoodsSyncRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = t.target.Scheme
	cloned.URL.Host = t.target.Host
	cloned.Host = t.target.Host
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(cloned)
}

func newProductGoodsSyncTestLogic(t *testing.T, core *app.Core, serverURL string) *ProductGoodsLogic {
	t.Helper()
	target, err := url.Parse(serverURL)
	require.NoError(t, err)
	logic := NewProductGoodsLogic(core)
	logic.httpClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: productGoodsSyncRewriteTransport{
			target: target,
			base:   http.DefaultTransport,
		},
	}
	return logic
}

func seedProductGoodsSyncTaxConfig(t *testing.T, core *app.Core) {
	t.Helper()
	_, err := core.DB().Exec(context.Background(), `
INSERT INTO system_config (config_key, config_value, description, created_at, updated_at)
VALUES
('finance_tax_exclusive_rate', '4.5', '未税->含税税率', ?, ?),
('finance_tax_inclusive_rate', '3.8', '含税->未税税率', ?, ?)
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), updated_at = VALUES(updated_at)
`, core.Now(), core.Now(), core.Now(), core.Now())
	require.NoError(t, err)
}

func seedProductGoodsSyncGoods(t *testing.T, core *app.Core, syncCost, syncName int, host string, goodsHasTax, channelHasTax int, goodsStatus ...int) int64 {
	t.Helper()
	status := 1
	if len(goodsStatus) > 0 {
		status = goodsStatus[0]
	}
	ctx := context.Background()
	now := core.Now()
	subjectRes, err := core.DB().Exec(ctx, `INSERT INTO admin_subject (name, has_tax, created_at, updated_at) VALUES (?, 1, ?, ?)`, "同步主体", now, now)
	require.NoError(t, err)
	subjectID, err := subjectRes.LastInsertId()
	require.NoError(t, err)
	brandRes, err := core.DB().Exec(ctx, `INSERT INTO product_brand (parent_id, name, icon, credential_image, is_visible, sort, goods_count, created_at, updated_at) VALUES (0, '同步品牌', '', '', 1, 1, 0, ?, ?)`, now, now)
	require.NoError(t, err)
	brandID, err := brandRes.LastInsertId()
	require.NoError(t, err)
	goodsRes, err := core.DB().Exec(ctx, `
INSERT INTO product_goods (
    goods_code, brand_id, name, goods_type, supply_type, is_export, is_douyin, has_tax, subject_id,
    exception_notify, purchase_notice, terminal_price_limit, balance_limit, default_sell_price,
    min_purchase_qty, max_purchase_qty, status, is_deleted, created_at, updated_at
) VALUES (?, ?, '同步商品', 'direct_recharge', 'channel', 1, 0, ?, ?, 1, '', NULL, '0.0000', '20.0000', 1, 1, ?, 0, ?, ?)
`, "SYNC001", brandID, goodsHasTax, subjectID, status, now, now)
	require.NoError(t, err)
	goodsID, err := goodsRes.LastInsertId()
	require.NoError(t, err)
	platformRes, err := core.DB().Exec(ctx, `
INSERT INTO supplier_platform_account (
    name, provider_code, provider_name, type_id, subject_id, has_tax, status, domain, backup_domain,
    token_id, secret_key, extra_config, threshold_amount, sort, crowd_name, is_deleted, created_at, updated_at
) VALUES ('同步卡卡云', 'kakayun', '卡卡云', 6, ?, ?, 1, ?, ?, 'merchant001', 'secretXYZ', '{}', '0.0000', 1, '', 0, ?, ?)
`, subjectID, channelHasTax, host, host, now, now)
	require.NoError(t, err)
	platformID, err := platformRes.LastInsertId()
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_config (
    goods_id, smart_reorder_enabled, reorder_timeout_enabled, reorder_timeout_minutes, order_strategy,
    sync_cost_price_enabled, sync_goods_name_enabled, allow_loss_sale_enabled, max_loss_amount, combo_goods_enabled,
    created_at, updated_at
) VALUES (?, 0, 0, 0, 'fixed_order', ?, ?, 0, '0.0000', 0, ?, ?)
`, goodsID, syncCost, syncName, now, now)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price,
    tax_adjust_direction, tax_adjust_rate, tax_adjust_amount, dock_status, sort, order_weight,
    order_time_start, order_time_end, validate_template_id, is_auto_change, add_type, default_price,
    is_deleted, created_at, updated_at
) VALUES (?, ?, 'SKU-100', '人工名称', '10.0000', '10.0000', 'none', '0.0000', '0.0000', 1, 1, '100.0000', NULL, NULL, NULL, 1, 'fixed', '0.2000', 0, ?, ?)
`, goodsID, platformID, now, now)
	require.NoError(t, err)
	return goodsID
}

func seedProductGoodsSyncExtraBinding(t *testing.T, core *app.Core, goodsID int64, supplierGoodsNo string) {
	t.Helper()
	ctx := context.Background()
	now := core.Now()
	platformIDValue, err := core.DB().GetCore().GetValue(ctx, `
SELECT platform_account_id
FROM product_goods_channel_binding
WHERE goods_id = ? AND is_deleted = 0
ORDER BY id ASC
LIMIT 1
`, goodsID)
	require.NoError(t, err)
	_, err = core.DB().Exec(ctx, `
INSERT INTO product_goods_channel_binding (
    goods_id, platform_account_id, supplier_goods_no, supplier_goods_name, source_cost_price, cost_price,
    tax_adjust_direction, tax_adjust_rate, tax_adjust_amount, dock_status, sort, order_weight,
    order_time_start, order_time_end, validate_template_id, is_auto_change, add_type, default_price,
    is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, '10.0000', '10.0000', 'none', '0.0000', '0.0000', 1, 1, '100.0000', NULL, NULL, NULL, 1, 'fixed', '0.2000', 0, ?, ?)
`, goodsID, platformIDValue.Int64(), supplierGoodsNo, "人工-"+supplierGoodsNo, now, now)
	require.NoError(t, err)
}

func loadProductGoodsSyncBinding(t *testing.T, core *app.Core, goodsID int64) productGoodsSyncBindingSnapshot {
	t.Helper()
	row := productGoodsSyncBindingSnapshot{}
	err := core.DB().GetCore().GetScan(context.Background(), &row, `
SELECT supplier_goods_name, source_cost_price, cost_price, tax_adjust_direction, tax_adjust_rate, tax_adjust_amount, is_auto_change, default_price
FROM product_goods_channel_binding
WHERE goods_id = ? AND is_deleted = 0
`, goodsID)
	require.NoError(t, err)
	return row
}

type productGoodsSyncBindingSnapshot struct {
	SupplierGoodsName  string `db:"supplier_goods_name"`
	SourceCostPrice    string `db:"source_cost_price"`
	CostPrice          string `db:"cost_price"`
	TaxAdjustDirection string `db:"tax_adjust_direction"`
	TaxAdjustRate      string `db:"tax_adjust_rate"`
	TaxAdjustAmount    string `db:"tax_adjust_amount"`
	IsAutoChange       int    `db:"is_auto_change"`
	DefaultPrice       string `db:"default_price"`
}
