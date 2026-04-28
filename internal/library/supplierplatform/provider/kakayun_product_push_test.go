package supplierprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKakayunProductSubscriptionProviderBuildRequests(t *testing.T) {
	provider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)

	account := AccountConfig{ProviderCode: "kakayun", TokenID: "10052", SecretKey: "secretXYZ"}
	now := time.Unix(1735002156, 0)

	subscribeReq, err := provider.BuildSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, subscribeReq.Method)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/subscribe", subscribeReq.URL.String())
	subscribeBody := decodeJSONBodyAny(t, readRequestBody(t, subscribeReq))
	require.Equal(t, "10052", subscribeBody["userid"])
	require.Equal(t, float64(1735002156), subscribeBody["timestamp"])
	require.Equal(t, "2582531", subscribeBody["goodsid"])
	require.NotEmpty(t, subscribeBody["sign"])
	require.NotContains(t, subscribeBody, "receiveurl")
	require.NotContains(t, subscribeBody, "oldreceiveurl")

	cancelReq, err := provider.BuildCancelSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/cancelsubscribe", cancelReq.URL.String())
	cancelBody := decodeJSONBodyAny(t, readRequestBody(t, cancelReq))
	require.Equal(t, "10052", cancelBody["userid"])
	require.Equal(t, float64(1735002156), cancelBody["timestamp"])
	require.Equal(t, "2582531", cancelBody["goodsid"])
	require.NotEmpty(t, cancelBody["sign"])
	require.NotContains(t, cancelBody, "receiveurl")
	require.NotContains(t, cancelBody, "oldreceiveurl")
}

func TestKakayunProductSubscriptionProviderParsesResponses(t *testing.T) {
	provider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)

	message, err := provider.ParseMutationResponse(http.StatusOK, []byte(`{"code":1,"msg":"成功"}`))
	require.NoError(t, err)
	require.Equal(t, "成功", message)

	_, err = provider.ParseMutationResponse(http.StatusOK, []byte(`{"code":0,"msg":"签名错误"}`))
	require.Error(t, err)
}

func TestKakayunProductChangePushProviderVerifyAndParse(t *testing.T) {
	provider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)

	account := AccountConfig{ProviderCode: "kakayun", SecretKey: "secretXYZ"}
	payload := map[string]any{
		"goodsid":     2582531,
		"goodsprice":  "52.9901",
		"goodsstock":  985,
		"goodsstatus": 1,
		"goodstype":   1,
		"goodsname":   "API直充接口测试",
		"update_time": 1735002156,
		"timestamp":   1735002156,
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	raw := marshalJSONForTest(t, payload)

	result, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), raw)
	require.NoError(t, err)
	require.Equal(t, "2582531", result.SupplierGoodsNo)
	require.Equal(t, "API直充接口测试", result.GoodsName)
	require.Equal(t, "52.9901", result.GoodsPrice.StringFixed(4))
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "1", result.GoodsStatus)
	require.Contains(t, result.Raw, "API直充接口测试")
}

func TestKakayunProductChangePushProviderRejectsInvalidSignAndExpiredTimestamp(t *testing.T) {
	provider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "kakayun", SecretKey: "secretXYZ"}

	payload := map[string]any{
		"goodsid":    "2582531",
		"goodsprice": "52.9901",
		"goodsname":  "API直充接口测试",
		"timestamp":  strconv.FormatInt(time.Unix(1735002156, 0).Unix(), 10),
		"sign":       "bad-sign",
	}
	_, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)

	payload["sign"] = kakayunSign(payload, account.SecretKey)
	_, err = provider.ParseProductChangePush(account, time.Unix(1735003000, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)
}

func TestLookupProductSubscriptionOnlyRegistersKakayun(t *testing.T) {
	subscriptionProvider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", subscriptionProvider.Code())

	for _, code := range []string{"kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"} {
		t.Run(code, func(t *testing.T) {
			_, ok := LookupProductSubscription(code)
			require.False(t, ok)
		})
	}
}

func TestLookupProductPushKeepsUnverifiableProvidersUnregistered(t *testing.T) {
	pushProvider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", pushProvider.Code())

	// 卡易信推送签名依赖请求头，当前 provider 接口只接收 body，不能验签时不注册。
	_, ok = LookupProductSubscription("kayixin")
	require.False(t, ok)
	_, ok = LookupProductChangePush("kayixin")
	require.False(t, ok)
}

func TestLookupProductChangePushRegistersSupportedNonKakayunProviders(t *testing.T) {
	for _, code := range []string{"kakayun", "kasushou", "xingquanyi", "youkayun"} {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupProductChangePush(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
	for _, code := range []string{"kayixin", "julangyun", "xinghai", "feisuyuan"} {
		t.Run(code, func(t *testing.T) {
			_, ok := LookupProductChangePush(code)
			require.False(t, ok)
		})
	}
}

func TestKasushouProductChangePushProviderParse(t *testing.T) {
	provider, ok := LookupProductChangePush("kasushou")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "kasushou", SecretKey: "secretXYZ"}
	payload := map[string]any{"id": "10001", "goods_price": "12.3400", "status": "1", "time": "1735002156123"}
	payload["sign"] = sha1Lower("1735002156123" + `{"id":"10001","time":"1735002156123"}` + "secretXYZ")
	result, err := provider.ParseProductChangePush(account, time.UnixMilli(1735002156123), marshalJSONForTest(t, payload))
	require.NoError(t, err)
	require.Equal(t, "10001", result.SupplierGoodsNo)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "12.3400", result.GoodsPrice.StringFixed(4))
	require.Equal(t, "1", result.GoodsStatus)
}

func TestXingquanyiProductChangePushProviderParseAndRejectsInvalidSign(t *testing.T) {
	provider, ok := LookupProductChangePush("xingquanyi")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "xingquanyi", SecretKey: "secretXYZ"}
	payload := map[string]any{
		"product_id":   "20001",
		"product_name": "星权益会员",
		"event_data": map[string]any{
			"price":  "15.6700",
			"status": "1",
		},
	}
	payload["sign"] = xingquanyiPushSign(payload, account.SecretKey)

	result, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.NoError(t, err)
	require.Equal(t, "20001", result.SupplierGoodsNo)
	require.Equal(t, "星权益会员", result.GoodsName)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "15.6700", result.GoodsPrice.StringFixed(4))
	require.Equal(t, "1", result.GoodsStatus)

	payload = map[string]any{
		"product_id":   "20001",
		"product_name": "星权益会员",
		"event_data":   `{"price":"18.9900","supply_state":1}`,
		"event_type":   "price_changed",
		"timestamp":    1735002160,
	}
	payload["sign"] = xingquanyiPushSign(payload, account.SecretKey)

	result, err = provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.NoError(t, err)
	require.Equal(t, "20001", result.SupplierGoodsNo)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "18.9900", result.GoodsPrice.StringFixed(4))
	require.Equal(t, "1", result.GoodsStatus)

	payload["sign"] = "bad-sign"
	_, err = provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)
}

func TestYoukayunProductChangePushProviderParseAndRejectsInvalidSign(t *testing.T) {
	provider, ok := LookupProductChangePush("youkayun")
	require.True(t, ok)
	account := AccountConfig{ProviderCode: "youkayun", SecretKey: "secretXYZ"}
	payload := map[string]any{
		"goods_id":    "30001",
		"goods_name":  "优卡云会员",
		"goods_price": "9.9900",
		"status":      1,
	}
	payload["sign"] = youkayunPushSign(payload, account.SecretKey)

	result, err := provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.NoError(t, err)
	require.Equal(t, "30001", result.SupplierGoodsNo)
	require.Equal(t, "优卡云会员", result.GoodsName)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "9.9900", result.GoodsPrice.StringFixed(4))
	require.Equal(t, "1", result.GoodsStatus)

	payload["sign"] = "bad-sign"
	_, err = provider.ParseProductChangePush(account, time.Unix(1735002160, 0), marshalJSONForTest(t, payload))
	require.Error(t, err)
}

func marshalJSONForTest(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}
