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

	getURLReq, err := provider.BuildGetReceiveURLsRequest(context.Background(), account, now)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, getURLReq.Method)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/user/geturl", getURLReq.URL.String())
	getURLBody := decodeJSONBodyAny(t, readRequestBody(t, getURLReq))
	require.Equal(t, "10052", getURLBody["userid"])
	require.Equal(t, float64(1735002156), getURLBody["timestamp"])
	require.NotEmpty(t, getURLBody["sign"])

	setURLReq, err := provider.BuildSetReceiveURLRequest(context.Background(), account, now, ProductReceiveURLInput{ReceiveURL: "https://example.com/callback"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/user/seturl", setURLReq.URL.String())
	setURLBody := decodeJSONBodyAny(t, readRequestBody(t, setURLReq))
	require.Equal(t, "https://example.com/callback", setURLBody["receiveurl"])

	subscribeReq, err := provider.BuildSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/subscribe", subscribeReq.URL.String())
	subscribeBody := decodeJSONBodyAny(t, readRequestBody(t, subscribeReq))
	require.Equal(t, "2582531", subscribeBody["goodsid"])

	cancelReq, err := provider.BuildCancelSubscribeRequest(context.Background(), account, now, ProductSubscribeInput{SupplierGoodsNo: "2582531"})
	require.NoError(t, err)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/cancelsubscribe", cancelReq.URL.String())
	cancelBody := decodeJSONBodyAny(t, readRequestBody(t, cancelReq))
	require.Equal(t, "2582531", cancelBody["goodsid"])
}

func TestKakayunProductSubscriptionProviderParsesResponses(t *testing.T) {
	provider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)

	urls, message, err := provider.ParseGetReceiveURLsResponse(http.StatusOK, []byte(`{"code":1,"msg":"success","data":[{"url":"https://a.example.com/cb","createtime":1740139648}]}`))
	require.NoError(t, err)
	require.Equal(t, "success", message)
	require.Len(t, urls, 1)
	require.Equal(t, "https://a.example.com/cb", urls[0].URL)
	require.Equal(t, int64(1740139648), urls[0].CreatedAtUnix)

	message, err = provider.ParseMutationResponse(http.StatusOK, []byte(`{"code":1,"msg":"成功"}`))
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

func TestLookupProductPushOnlyRegistersKakayun(t *testing.T) {
	subscriptionProvider, ok := LookupProductSubscription("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", subscriptionProvider.Code())

	pushProvider, ok := LookupProductChangePush("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", pushProvider.Code())

	_, ok = LookupProductSubscription("kayixin")
	require.False(t, ok)
	_, ok = LookupProductChangePush("kayixin")
	require.False(t, ok)
}

func marshalJSONForTest(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}
