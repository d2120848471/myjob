package supplierprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProvidersBuildRequest(t *testing.T) {
	now := time.Date(2026, 4, 14, 12, 30, 45, 678000000, time.UTC)
	baseURL := "http://platform.example.com"
	account := AccountConfig{
		Domain:       "main.example.com",
		BackupDomain: "backup.example.com",
		TokenID:      "merchant001",
		SecretKey:    "secretXYZ",
		ExtraConfig:  map[string]any{},
	}

	cases := []struct {
		name     string
		provider BalanceProvider
		account  AccountConfig
		baseURL  string
		assert   func(t *testing.T, req *http.Request, body []byte)
	}{
		{
			name:     "kakayun",
			provider: kakayunProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/dockapiv3/user/info", req.URL.String())
				require.Equal(t, "curl/7.81.0", req.Header.Get("User-Agent"))
				expected := md5Lower("timestamp=" + strconv.FormatInt(now.Unix(), 10) + "&userid=merchant001" + "secretXYZ")
				require.Equal(t, `{"userid":"merchant001","timestamp":`+strconv.FormatInt(now.Unix(), 10)+`,"sign":"`+expected+`"}`, string(body))
				payload := decodeJSONBodyAny(t, body)
				require.Equal(t, "merchant001", payload["userid"])
				require.Equal(t, float64(now.Unix()), payload["timestamp"])
				require.Equal(t, expected, payload["sign"])
			},
		},
		{
			name:     "kayixin",
			provider: kayixinProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/api/v3/user/getAccount", req.URL.String())
				require.Equal(t, "application/json", req.Header.Get("Content-Type"))
				require.Empty(t, string(body))
				timestamp := strconv.FormatInt(now.UnixMilli(), 10)
				require.Equal(t, "3.0", req.Header.Get("X-Version"))
				require.Equal(t, "merchant001", req.Header.Get("X-App-Id"))
				require.Equal(t, timestamp, req.Header.Get("X-Timestamp"))
				expected := md5Lower("merchant001" + "secretXYZ" + "3.0" + timestamp)
				require.Equal(t, expected, req.Header.Get("X-Signature"))
			},
		},
		{
			name:     "kasushou",
			provider: kasushouProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/api/v1/user/info", req.URL.String())
				require.Equal(t, "application/json", req.Header.Get("Content-Type"))
				require.Equal(t, "{}", string(body))
				timestamp := strconv.FormatInt(now.UnixMilli(), 10)
				require.Equal(t, "merchant001", req.Header.Get("UserId"))
				require.Equal(t, timestamp, req.Header.Get("Timestamp"))
				expected := sha1Lower(timestamp + "{}" + "secretXYZ")
				require.Equal(t, expected, req.Header.Get("Sign"))
			},
		},
		{
			name:     "xingquanyi",
			provider: xingquanyiProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/api/customer", req.URL.String())
				payload := decodeJSONBody(t, body)
				require.Equal(t, "merchant001", payload["customer_id"])
				require.Equal(t, strconv.FormatInt(now.Unix(), 10), payload["timestamp"])
				expected := md5Lower("secretXYZ" + "customer_id" + "merchant001" + "timestamp" + strconv.FormatInt(now.Unix(), 10))
				require.Equal(t, expected, payload["sign"])
			},
		},
		{
			name:     "youkayun",
			provider: youkayunProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/api/getusermoney", req.URL.String())
				fields := decodeMultipartFields(t, req, body)
				require.Equal(t, "merchant001", fields["userid"])
				expected := md5Lower("userid=merchant001" + "secretXYZ")
				require.Equal(t, expected, fields["sign"])
			},
		},
		{
			name:     "julangyun",
			provider: julangyunProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodGet, req.Method)
				require.Equal(t, "http://platform.example.com/api/recharge/user/amount/detail", req.URL.String())
				require.Empty(t, body)
				timestamp := strconv.FormatInt(now.UnixMilli(), 10)
				require.Equal(t, "merchant001", req.Header.Get("userCode"))
				require.Equal(t, timestamp, req.Header.Get("timestamp"))
				expected := md5Lower("timestamp" + timestamp + "userCode" + "merchant001" + "secretXYZ")
				require.Equal(t, expected, req.Header.Get("sign"))
			},
		},
		{
			name:     "feisuyuan",
			provider: feisuyuanProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/recharge/info", req.URL.String())
				values, err := url.ParseQuery(string(body))
				require.NoError(t, err)
				require.Equal(t, "merchant001", values.Get("merchantId"))
				require.Equal(t, strconv.FormatInt(now.Unix(), 10), values.Get("timeStamp"))
				require.Equal(t, "1.0", values.Get("version"))
				expected := md5Upper("merchantId=merchant001&timeStamp=" + strconv.FormatInt(now.Unix(), 10) + "&key=secretXYZ")
				require.Equal(t, expected, values.Get("sign"))
			},
		},
		{
			name:     "xinghai",
			provider: xinghaiProvider{},
			account:  account,
			baseURL:  baseURL,
			assert: func(t *testing.T, req *http.Request, body []byte) {
				require.Equal(t, http.MethodPost, req.Method)
				require.Equal(t, "http://platform.example.com/api/account/balance", req.URL.String())
				values, err := url.ParseQuery(string(body))
				require.NoError(t, err)
				timestamp := now.Format("20060102150405000")
				require.Equal(t, "merchant001", values.Get("appId"))
				require.Equal(t, timestamp, values.Get("timestamp"))
				expected := md5Lower("appId=merchant001&appSecret=secretXYZ&timestamp=" + timestamp)
				require.Equal(t, expected, values.Get("sign"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.provider.BuildRequest(context.Background(), tc.account, now, tc.baseURL)
			require.NoError(t, err)
			body := readRequestBody(t, req)
			tc.assert(t, req, body)
		})
	}
}

func TestKakayunCandidateBaseURLs_UsesConfiguredDomains(t *testing.T) {
	account := AccountConfig{
		Domain:       "main.example.com",
		BackupDomain: "backup.example.com",
	}

	require.Equal(t,
		[]string{
			"http://main.example.com",
			"https://main.example.com",
			"http://backup.example.com",
			"https://backup.example.com",
		},
		kakayunProvider{}.CandidateBaseURLs(account),
	)
}

func TestOrderProviderCapabilities(t *testing.T) {
	tests := []struct {
		code     string
		provider interface {
			Capabilities() OrderProviderCapabilities
		}
		maxQty      int
		safetyMode  SafetyPriceMode
		safetyField string
	}{
		{code: "kakayun", provider: kakayunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "maxmoney"},
		{code: "kayixin", provider: kayixinProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "safePrice"},
		{code: "kasushou", provider: kasushouProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "safe_price"},
		{code: "xingquanyi", provider: xingquanyiProvider{}, maxQty: 0, safetyMode: SafetyPriceModeUnit, safetyField: "safe_cost"},
		{code: "youkayun", provider: youkayunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "maxmoney"},
		{code: "julangyun", provider: julangyunProvider{}, maxQty: 0, safetyMode: SafetyPriceModeTotal, safetyField: "accessPrice"},
		{code: "xinghai", provider: xinghaiProvider{}, maxQty: 0, safetyMode: SafetyPriceModeUnit, safetyField: "itemPrice"},
		{code: "feisuyuan", provider: feisuyuanProvider{}, maxQty: 1, safetyMode: SafetyPriceModeUnsupported, safetyField: ""},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			capabilities := tc.provider.Capabilities()
			require.Equal(t, tc.maxQty, capabilities.MaxQuantityPerCreate)
			require.Equal(t, tc.safetyMode, capabilities.SafetyPrice.Mode)
			require.Equal(t, tc.safetyField, capabilities.SafetyPrice.FieldName)
		})
	}
}

func TestKakayunProductInfoProviderBuildRequestAndParse(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)

	account := AccountConfig{
		ProviderCode: "kakayun",
		Domain:       "qqlogin.yxp8.cn",
		TokenID:      "10052",
		SecretKey:    "9aa3034b6beba7cf5bfcf6089218a674",
	}
	now := time.Unix(1735002156, 0)
	req, err := provider.BuildProductInfoRequest(context.Background(), account, now, "http://qqlogin.yxp8.cn", ProductInfoInput{
		SupplierGoodsNo: "2478510",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "http://public.kky.v3.api.kakayun.vip/dockapiv3/goods/details", req.URL.String())
	require.Equal(t, "curl/7.81.0", req.Header.Get("User-Agent"))

	body := readRequestBody(t, req)
	payload := decodeJSONBodyAny(t, body)
	require.Equal(t, "10052", payload["userid"])
	require.Equal(t, float64(1735002156), payload["timestamp"])
	require.Equal(t, "2478510", payload["goodsid"])
	require.NotEmpty(t, payload["sign"])

	result, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(`{"code":1,"message":"ok","data":{"goodsid":"2478510","goodsname":"测试产品","goodsprice":"11","stock":9999,"goodsstatus":1}}`))
	require.NoError(t, err)
	require.Equal(t, "2478510", result.SupplierGoodsNo)
	require.Equal(t, "测试产品", result.GoodsName)
	require.True(t, result.GoodsPriceValid)
	require.Equal(t, "11.0000", result.GoodsPrice.StringFixed(4))
	require.Contains(t, result.Raw, `"goodsname":"测试产品"`)
}

func TestKakayunProductInfoProviderParsesNumericGoodsID(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)

	result, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(`{"code":1,"msg":"success","data":{"goodsid":2582531,"goodsname":"测试产品888","goodsprice":"0.1000","goodsstatus":1}}`))
	require.NoError(t, err)
	require.Equal(t, "2582531", result.SupplierGoodsNo)
	require.Equal(t, "测试产品888", result.GoodsName)
	require.Equal(t, "0.1000", result.GoodsPrice.StringFixed(4))
}

func TestKakayunProductInfoProviderRejectsInvalidResponses(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)

	tests := []struct {
		name string
		body string
	}{
		{name: "非 JSON", body: `<html>bad gateway</html>`},
		{name: "业务失败", body: `{"code":0,"message":"商品不存在"}`},
		{name: "缺少数据", body: `{"code":1,"message":"ok"}`},
		{name: "没有可同步字段", body: `{"code":1,"data":{"goodsid":"2478510","goodsname":"","goodsprice":"abc"}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(tc.body))
			require.Error(t, err)
		})
	}
}

func TestKakayunProductInfoProviderRejectsHTTPErrorStatus(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)

	_, err := provider.ParseProductInfoResponse(http.StatusBadGateway, []byte(`{"code":1,"data":{"goodsid":"2478510","goodsname":"测试产品","goodsprice":"11"}}`))
	require.Error(t, err)
}

func TestKakayunProductInfoProviderKeepsNameWhenPriceInvalid(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)

	result, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(`{"code":1,"data":{"goodsid":"2478510","goodsname":"测试产品","goodsprice":"abc"}}`))
	require.NoError(t, err)
	require.Equal(t, "2478510", result.SupplierGoodsNo)
	require.Equal(t, "测试产品", result.GoodsName)
	require.False(t, result.GoodsPriceValid)
}

func TestLookupProductInfoIncludesKakayun(t *testing.T) {
	provider, ok := LookupProductInfo("kakayun")
	require.True(t, ok)
	require.Equal(t, "kakayun", provider.Code())
}

func TestProductInfoProviderRegistriesIncludeAllConfiguredPlatforms(t *testing.T) {
	codes := []string{"kakayun", "kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupProductInfo(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
}

func TestMultiPlatformProductInfoProvidersBuildAndParse(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 0, time.UTC)
	account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: map[string]any{}}
	tests := []struct {
		code string
		path string
		body string
	}{
		{"kayixin", "/api/v3/goods/getDetail", `{"code":1000,"msg":"success","data":{"goodsId":10001,"name":"卡易信会员","salesPrice":12.34,"status":1}}`},
		{"kasushou", "/api/v1/goods/info", `{"code":200,"msg":"成功","data":{"id":10001,"goods_name":"卡速售会员","goods_price":"23.4500","status":1}}`},
		{"xingquanyi", "/api/product", `{"code":"ok","message":"","data":{"id":10001,"product_name":"星权益会员","name":"月卡","price":"34.5600"}}`},
		{"youkayun", "/api/goodsdetails", `{"code":1000,"msg":"查询成功","data":{"id":10001,"goods_name":"优卡云会员","goods_price":"45.6700","status":1}}`},
		{"julangyun", "/api/recharge/goods/detail", `{"code":200,"message":"处理成功","data":{"goodsCode":"10001","goodsName":"聚浪云会员","goodsPrice":56.78,"goodsStatus":1}}`},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupProductInfo(tc.code)
			require.True(t, ok)
			req, err := provider.BuildProductInfoRequest(context.Background(), account, now, "http://platform.example.com", ProductInfoInput{SupplierGoodsNo: "10001"})
			require.NoError(t, err)
			require.Equal(t, "http://platform.example.com"+tc.path, req.URL.String())

			info, err := provider.ParseProductInfoResponse(http.StatusOK, []byte(tc.body))
			require.NoError(t, err)
			require.Equal(t, "10001", info.SupplierGoodsNo)
			require.NotEmpty(t, info.GoodsName)
			require.True(t, info.GoodsPriceValid)
		})
	}
}

func TestListBasedProductInfoProvidersParseMatchedItem(t *testing.T) {
	tests := []struct {
		code  string
		body  string
		name  string
		price string
	}{
		{"xinghai", `{"code":"00","msg":"查询成功","data":[{"itemId":"10001","itemName":"星海会员","price":"67.8900"},{"itemId":"10002","itemName":"其他","price":"1"}]}`, "星海会员", "67.8900"},
		{"feisuyuan", `{"code":"0000","products":[{"product_id":"10001","channel_price":"78.9000","item_name":"飞速源会员"},{"product_id":"10002","channel_price":"1","item_name":"其他"}]}`, "飞速源会员", "78.9000"},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupProductInfo(tc.code)
			require.True(t, ok)
			listProvider, ok := provider.(ProductInfoListProvider)
			require.True(t, ok)
			results, err := listProvider.ParseProductInfoListResponse(http.StatusOK, []byte(tc.body))
			require.NoError(t, err)
			require.Equal(t, tc.name, results["10001"].GoodsName)
			require.Equal(t, tc.price, results["10001"].GoodsPrice.StringFixed(4))
		})
	}
}

func TestKayixinBuildRequest_UsesObjectBodyWhenConfigured(t *testing.T) {
	now := time.Date(2026, 4, 14, 12, 30, 45, 678000000, time.UTC)
	req, err := kayixinProvider{}.BuildRequest(context.Background(), AccountConfig{
		TokenID:     "merchant001",
		SecretKey:   "secretXYZ",
		ExtraConfig: map[string]any{"sign_with_empty_object": true},
	}, now, "http://platform.example.com")
	require.NoError(t, err)

	body := string(readRequestBody(t, req))
	require.Equal(t, "{}", body)
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	expected := md5Lower("merchant001" + "secretXYZ" + "3.0" + timestamp + "{}")
	require.Equal(t, expected, req.Header.Get("X-Signature"))
}

func TestKakayunOrderProviderBuildCreateRequest(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)
	account := AccountConfig{
		ProviderCode: "kakayun",
		Domain:       "qqlogin.yxp8.cn",
		TokenID:      "10052",
		SecretKey:    "9aa3034b6beba7cf5bfcf6089218a674",
	}
	req, err := provider.BuildCreateOrderRequest(context.Background(), account, time.Unix(1735002156, 0), "http://qqlogin.yxp8.cn", CreateOrderInput{
		SupplierGoodsNo:   "720938",
		Quantity:          1,
		Account:           "13088888888",
		SupplierUSOrderNo: "O20260424153000123456-T1",
		MaxMoney:          "11.0000",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "http://qqlogin.yxp8.cn/dockapiv3/order/create", req.URL.String())
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Equal(t, "10052", payload["userid"])
	require.Equal(t, float64(1735002156), payload["timestamp"])
	require.Equal(t, "720938", payload["goodsid"])
	require.Equal(t, float64(1), payload["buynum"])
	require.Equal(t, "13088888888", payload["attach"])
	require.Equal(t, "O20260424153000123456-T1", payload["usorderno"])
	require.Equal(t, "11.0000", payload["maxmoney"])

	require.Equal(t, "f61433b61b3d2fed6cbcd9cc9ca147dd", payload["sign"])
}

func TestKakayunOrderProviderParseCreateAndQuery(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)

	create, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":1,"message":"下单成功","data":{"orderno":"SD202604240001","usorderno":"O20260424153000123456-T1"}}`))
	require.NoError(t, err)
	require.True(t, create.Accepted)
	require.Equal(t, "SD202604240001", create.SupplierOrderNo)
	require.Equal(t, "O20260424153000123456-T1", create.SupplierUSOrderNo)

	processing, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1,"message":"ok","data":{"orderno":"SD202604240001","usorderno":"O20260424153000123456-T1","status":3,"refundstatus":0,"receipt":"处理中"}}`))
	require.NoError(t, err)
	require.Equal(t, SupplierOrderStatusProcessing, processing.Status)

	success, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1,"message":"ok","data":{"orderno":"SD202604240001","usorderno":"O20260424153000123456-T1","status":5,"refundstatus":0,"receipt":"成功"}}`))
	require.NoError(t, err)
	require.Equal(t, SupplierOrderStatusSuccess, success.Status)

	failed, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1,"message":"ok","data":{"orderno":"SD202604240001","usorderno":"O20260424153000123456-T1","status":4,"refundstatus":1,"receipt":"失败"}}`))
	require.NoError(t, err)
	require.Equal(t, SupplierOrderStatusFailed, failed.Status)
}

func TestKakayunOrderProviderNonJSONIsUnknown(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)
	result, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`<html>bad gateway</html>`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSupplierUnknownResponse)
	require.Equal(t, SupplierOrderStatusUnknown, result.Status)
}

func TestKakayunOrderProviderCreateCode9999IsUnknown(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)

	result, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":9999,"message":"系统繁忙","data":{"usorderno":"O20260424153000123456-T1"}}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSupplierUnknownResponse)
	require.False(t, result.Accepted)
	require.Equal(t, SupplierOrderStatusUnknown, result.Status)
	require.Equal(t, "9999", result.SupplierStatus)
	require.Equal(t, "系统繁忙", result.Message)
	require.Equal(t, "O20260424153000123456-T1", result.SupplierUSOrderNo)
}

func TestKakayunOrderProviderCreateExplicitFailureIsFailed(t *testing.T) {
	provider, ok := LookupOrder("kakayun")
	require.True(t, ok)

	result, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":0,"message":"库存不足"}`))
	require.Error(t, err)
	require.False(t, result.Accepted)
	require.Equal(t, SupplierOrderStatusFailed, result.Status)
	require.Equal(t, "库存不足", result.Message)
}

func TestOrderProviderRegistriesIncludeAllConfiguredPlatforms(t *testing.T) {
	codes := []string{"kakayun", "kayixin", "kasushou", "xingquanyi", "youkayun", "julangyun", "xinghai", "feisuyuan"}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			provider, ok := LookupOrder(code)
			require.True(t, ok)
			require.Equal(t, code, provider.Code())
		})
	}
}

func TestMultiPlatformOrderProvidersBuildCreateRequests(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 123000000, time.UTC)
	account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: map[string]any{}}
	input := CreateOrderInput{SupplierGoodsNo: "10001", Quantity: 2, Account: "13800138000", SupplierUSOrderNo: "O20260428123045123456-T1-S1", SafePrice: "20.0000"}
	tests := []struct {
		code   string
		path   string
		assert func(t *testing.T, req *http.Request, body []byte)
	}{
		{"kayixin", "/api/v3/order/create", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, float64(10001), payload["goodsId"])
			require.Equal(t, float64(2), payload["count"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["outerNumber"])
			require.Equal(t, float64(20), payload["safePrice"])
			require.Equal(t, []any{map[string]any{"name": "充值账号", "value": "13800138000"}}, payload["attach"])
		}},
		{"kasushou", "/api/v1/order/buy", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, float64(10001), payload["id"])
			require.Equal(t, float64(2), payload["quantity"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["external_orderno"])
			require.Equal(t, "20.0000", payload["safe_price"])
			require.Equal(t, map[string]any{"recharge_account": "13800138000"}, payload["attach"])
		}},
		{"xingquanyi", "/api/buy", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, float64(10001), payload["product_id"])
			require.Equal(t, float64(2), payload["quantity"])
			require.Equal(t, "13800138000", payload["recharge_account"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["outer_order_id"])
			require.Equal(t, "20.0000", payload["safe_cost"])
			require.NotEmpty(t, payload["sign"])
		}},
		{"youkayun", "/api/buygoods", func(t *testing.T, req *http.Request, body []byte) {
			fields := decodeMultipartFields(t, req, body)
			require.Equal(t, "10001", fields["goodsid"])
			require.Equal(t, "2", fields["quantity"])
			require.Equal(t, "13800138000", fields["accountname"])
			require.Equal(t, "O20260428123045123456-T1-S1", fields["outorderno"])
			require.Equal(t, "20.0000", fields["maxmoney"])
		}},
		{"julangyun", "/api/recharge/order/submit", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, "10001", payload["goodsCode"])
			require.Equal(t, "O20260428123045123456-T1-S1", payload["accessOrderNo"])
			require.Equal(t, "13800138000", payload["rechargeAccount"])
			require.Equal(t, float64(2), payload["orderNum"])
			require.Equal(t, "", payload["callbackUrl"])
			require.NotContains(t, payload, "notifyUrl")
			require.Equal(t, float64(20), payload["accessPrice"])
		}},
		{"xinghai", "/api/order/submit", func(t *testing.T, req *http.Request, body []byte) {
			values, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			require.Equal(t, "10001", values.Get("itemId"))
			require.Equal(t, "2", values.Get("amount"))
			require.Equal(t, "13800138000", values.Get("uuid"))
			require.Equal(t, "O20260428123045123456-T1-S1", values.Get("outOrderId"))
			require.Equal(t, "20", values.Get("itemPrice"))
		}},
		{"feisuyuan", "/recharge/order", func(t *testing.T, req *http.Request, body []byte) {
			values, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			require.Equal(t, "10001", values.Get("productId"))
			require.Equal(t, "1", values.Get("number"))
			require.Equal(t, "0", values.Get("accountType"))
			require.Equal(t, "13800138000", values.Get("rechargeAccount"))
			require.Equal(t, "O20260428123045123456-T1-S1", values.Get("outTradeNo"))
			require.Empty(t, values.Get("maxmoney"))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupOrder(tc.code)
			require.True(t, ok)
			req, err := provider.BuildCreateOrderRequest(context.Background(), account, now, "http://platform.example.com", input)
			require.NoError(t, err)
			require.Equal(t, "http://platform.example.com"+tc.path, req.URL.String())
			tc.assert(t, req, readRequestBody(t, req))
		})
	}
}

func TestMultiPlatformOrderProvidersBuildQueryRequests(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 123000000, time.UTC)
	account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: map[string]any{}}
	input := QueryOrderInput{SupplierOrderNo: "SUP001", SupplierUSOrderNo: "OUT001"}
	tests := []struct {
		code   string
		path   string
		assert func(t *testing.T, req *http.Request, body []byte)
	}{
		{"kasushou", "/api/v1/order/info", func(t *testing.T, req *http.Request, body []byte) {
			payload := decodeJSONBodyAny(t, body)
			require.Equal(t, "OUT001", payload["external_orderno"])
			require.Equal(t, "SUP001", payload["ordersn"])
			require.Equal(t, "0", payload["day"])
		}},
		{"youkayun", "/api/queryorder", func(t *testing.T, req *http.Request, body []byte) {
			fields := decodeMultipartFields(t, req, body)
			require.Equal(t, "SUP001", fields["orderno"])
			require.Equal(t, "OUT001", fields["outer_order_id"])
			require.Equal(t, "merchant001", fields["userid"])
			require.NotEmpty(t, fields["sign"])
			require.NotContains(t, fields, "ordersn")
			require.NotContains(t, fields, "outorderno")
		}},
		{"feisuyuan", "/recharge/query", func(t *testing.T, req *http.Request, body []byte) {
			values, err := url.ParseQuery(string(body))
			require.NoError(t, err)
			require.Equal(t, "merchant001", values.Get("merchantId"))
			require.Equal(t, "OUT001", values.Get("outTradeNo"))
			require.Equal(t, strconv.FormatInt(now.Unix(), 10), values.Get("timeStamp"))
			require.Equal(t, "1.0", values.Get("version"))
			expected := md5Upper("merchantId=merchant001&outTradeNo=OUT001&timeStamp=" + strconv.FormatInt(now.Unix(), 10) + "&key=secretXYZ")
			require.Equal(t, expected, values.Get("sign"))
		}},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupOrder(tc.code)
			require.True(t, ok)
			req, err := provider.BuildQueryOrderRequest(context.Background(), account, now, "http://platform.example.com", input)
			require.NoError(t, err)
			require.Equal(t, "http://platform.example.com"+tc.path, req.URL.String())
			tc.assert(t, req, readRequestBody(t, req))
		})
	}
}

func TestMultiPlatformOrderProvidersParseCreateAndQuery(t *testing.T) {
	tests := []struct {
		code          string
		createBody    string
		querySuccess  string
		queryFailed   string
		queryProgress string
	}{
		{"kayixin", `{"code":1000,"msg":"购买成功","data":{"orderNumber":"KYX001"}}`, `{"code":1000,"data":{"orderNumber":"KYX001","outerNumber":"OUT001","status":3,"result":"完成"}}`, `{"code":1000,"data":{"status":4,"result":"退款"}}`, `{"code":1000,"data":{"status":2,"result":"处理中"}}`},
		{"kasushou", `{"code":200,"msg":"下单成功","data":{"ordersn":"KSS001","external_orderno":"OUT001"}}`, `{"code":200,"data":[{"ordersn":"KSS001","external_orderno":"OUT001","status":3,"recharge_hints":"完成"}]}`, `{"code":200,"data":[{"status":4,"recharge_hints":"取消"}]}`, `{"code":200,"data":[{"status":2,"recharge_hints":"处理中"}]}`},
		{"xingquanyi", `{"code":"ok","message":"","data":{"order_id":"XQY001","state":101}}`, `{"code":"ok","data":{"id":"XQY001","outer_order_id":"OUT001","state":200,"recharge_info":"完成"}}`, `{"code":"ok","data":{"state":500,"recharge_info":"失败"}}`, `{"code":"ok","data":{"state":101,"recharge_info":"处理中"}}`},
		{"youkayun", `{"code":1000,"msg":"获取成功","data":{"ordersn":"YKY001","outorderno":"OUT001"}}`, `{"code":1000,"data":{"ordersn":"YKY001","outer_order_no":"OUT001","status":3}}`, `{"code":1000,"data":{"status":5}}`, `{"code":1000,"data":{"status":2}}`},
		{"julangyun", `{"code":200,"message":"处理成功","data":{"returnOrderNo":"JLY001","accessOrderNo":"OUT001","orderStatus":20}}`, `{"code":200,"data":{"returnOrderNo":"JLY001","accessOrderNo":"OUT001","orderStatus":30}}`, `{"code":200,"data":{"orderStatus":40}}`, `{"code":200,"data":{"orderStatus":20}}`},
		{"xinghai", `{"code":"00","msg":"下单成功","orderId":"XH001","outOrderId":"OUT001"}`, `{"code":"00","orderId":"XH001","outOrderId":"OUT001","orderStatus":"2","orderDesc":"成功"}`, `{"code":"00","orderStatus":"3","orderDesc":"失败"}`, `{"code":"00","orderStatus":"1","orderDesc":"处理中"}`},
		{"feisuyuan", `{"code":"2000","message":"ok"}`, `{"code":"0000","status":"01","message":"success","outTradeNo":"OUT001"}`, `{"code":"0000","status":"03","message":"fail","outTradeNo":"OUT001"}`, `{"code":"0000","status":"02","message":"pending","outTradeNo":"OUT001"}`},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			provider, ok := LookupOrder(tc.code)
			require.True(t, ok)
			create, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(tc.createBody))
			require.NoError(t, err)
			require.True(t, create.Accepted)
			require.Equal(t, SupplierOrderStatusProcessing, create.Status)

			success, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.querySuccess))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusSuccess, success.Status)

			failed, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.queryFailed))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusFailed, failed.Status)

			processing, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(tc.queryProgress))
			require.NoError(t, err)
			require.Equal(t, SupplierOrderStatusProcessing, processing.Status)
		})
	}
}

func TestXingquanyiServerErrorIsUnknown(t *testing.T) {
	provider := xingquanyiProvider{}

	create, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":"server_error","message":"服务端未知错误"}`))
	require.Error(t, err)
	require.False(t, create.Accepted)
	require.Equal(t, SupplierOrderStatusUnknown, create.Status)
	require.NotEqual(t, SupplierOrderStatusFailed, create.Status)
	require.Equal(t, "server_error", create.SupplierStatus)
	require.Equal(t, "服务端未知错误", create.Message)

	query, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":"server_error","message":"服务端未知错误"}`))
	require.Error(t, err)
	require.Equal(t, SupplierOrderStatusUnknown, query.Status)
	require.NotEqual(t, SupplierOrderStatusFailed, query.Status)
	require.Equal(t, "server_error", query.SupplierStatus)
	require.Equal(t, "服务端未知错误", query.Message)
}

func TestFormatXinghaiItemPrice(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "20", want: "20"},
		{input: "20.1", want: "20.1"},
		{input: "20.1200", want: "20.12"},
		{input: "20.1234", want: "20.123"},
		{input: "20.1235", want: "20.124"},
		{input: "bad", want: "0"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			require.Equal(t, tc.want, formatXinghaiItemPrice(tc.input))
		})
	}
}

func TestFeisuyuanBuildCreateRequestAccountType(t *testing.T) {
	now := time.Date(2026, 4, 28, 12, 30, 45, 123000000, time.UTC)
	input := CreateOrderInput{SupplierGoodsNo: "10001", Quantity: 1, Account: "13800138000", SupplierUSOrderNo: "OUT001"}
	tests := []struct {
		name        string
		extraConfig map[string]any
		want        string
	}{
		{name: "default", extraConfig: nil, want: "0"},
		{name: "camel key", extraConfig: map[string]any{"accountType": 1}, want: "1"},
		{name: "snake key", extraConfig: map[string]any{"account_type": 2}, want: "2"},
		{name: "invalid", extraConfig: map[string]any{"accountType": 9}, want: "0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			account := AccountConfig{TokenID: "merchant001", SecretKey: "secretXYZ", ExtraConfig: tc.extraConfig}
			req, err := feisuyuanProvider{}.BuildCreateOrderRequest(context.Background(), account, now, "http://platform.example.com", input)
			require.NoError(t, err)
			values, err := url.ParseQuery(string(readRequestBody(t, req)))
			require.NoError(t, err)
			require.Equal(t, tc.want, values.Get("accountType"))
		})
	}
}

func TestProvidersParseBalanceResponse(t *testing.T) {
	cases := []struct {
		name           string
		provider       BalanceProvider
		successBody    string
		successAmount  string
		successMessage string
		failBody       string
		failMessage    string
		invalidBody    string
	}{
		{
			name:           "kakayun",
			provider:       kakayunProvider{},
			successBody:    `{"code":"1","msg":"查询成功","data":{"money":"12.3400"}}`,
			successAmount:  "12.3400",
			successMessage: "查询成功",
			failBody:       `{"code":"0","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"1","data":{}}`,
		},
		{
			name:           "kayixin",
			provider:       kayixinProvider{},
			successBody:    `{"code":"1000","message":"查询成功","data":{"balance":"23.4500"}}`,
			successAmount:  "23.4500",
			successMessage: "查询成功",
			failBody:       `{"code":"1001","message":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"1000","data":{}}`,
		},
		{
			name:           "kasushou",
			provider:       kasushouProvider{},
			successBody:    `{"code":"200","msg":"查询成功","data":{"balance":"34.5600"}}`,
			successAmount:  "34.5600",
			successMessage: "查询成功",
			failBody:       `{"code":"500","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"200","data":{}}`,
		},
		{
			name:           "xingquanyi",
			provider:       xingquanyiProvider{},
			successBody:    `{"code":"ok","msg":"查询成功","data":{"balance":"45.6700"}}`,
			successAmount:  "45.6700",
			successMessage: "查询成功",
			failBody:       `{"code":"invalid_sign","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"ok","data":{}}`,
		},
		{
			name:           "youkayun",
			provider:       youkayunProvider{},
			successBody:    `{"code":"1000","msg":"查询成功","money":"56.7800"}`,
			successAmount:  "56.7800",
			successMessage: "查询成功",
			failBody:       `{"code":"1001","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"1000"}`,
		},
		{
			name:           "julangyun",
			provider:       julangyunProvider{},
			successBody:    `{"code":"200","message":"查询成功","data":{"amount":"67.8900"}}`,
			successAmount:  "67.8900",
			successMessage: "查询成功",
			failBody:       `{"code":"500","message":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"200","data":{}}`,
		},
		{
			name:           "feisuyuan",
			provider:       feisuyuanProvider{},
			successBody:    `{"code":"0000","msg":"查询成功","balance":"78.9000"}`,
			successAmount:  "78.9000",
			successMessage: "查询成功",
			failBody:       `{"code":"9999","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"0000"}`,
		},
		{
			name:           "xinghai",
			provider:       xinghaiProvider{},
			successBody:    `{"code":"00","msg":"查询成功","balance":"89.0100"}`,
			successAmount:  "89.0100",
			successMessage: "查询成功",
			failBody:       `{"code":"99","msg":"签名错误"}`,
			failMessage:    "签名错误",
			invalidBody:    `{"code":"00"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"_success", func(t *testing.T) {
			amount, message, err := tc.provider.ParseBalanceResponse(http.StatusOK, []byte(tc.successBody))
			require.NoError(t, err)
			require.Equal(t, tc.successAmount, amount.StringFixed(4))
			require.Equal(t, tc.successMessage, message)
		})

		t.Run(tc.name+"_business_fail", func(t *testing.T) {
			amount, message, err := tc.provider.ParseBalanceResponse(http.StatusOK, []byte(tc.failBody))
			require.Error(t, err)
			require.True(t, amount.IsZero())
			require.Equal(t, tc.failMessage, message)
		})

		t.Run(tc.name+"_invalid", func(t *testing.T) {
			amount, message, err := tc.provider.ParseBalanceResponse(http.StatusOK, []byte(tc.invalidBody))
			require.Error(t, err)
			require.True(t, amount.IsZero())
			require.Equal(t, "余额解析失败", message)
		})
	}
}

func readRequestBody(t *testing.T, req *http.Request) []byte {
	t.Helper()
	if req.Body == nil {
		return nil
	}
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	return body
}

func decodeJSONBody(t *testing.T, body []byte) map[string]string {
	t.Helper()
	payload := make(map[string]string)
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload
}

func decodeJSONBodyAny(t *testing.T, body []byte) map[string]any {
	t.Helper()
	payload := make(map[string]any)
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload
}

func decodeMultipartFields(t *testing.T, req *http.Request, body []byte) map[string]string {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(mediaType, "multipart/form-data"))
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	fields := make(map[string]string)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		data, readErr := io.ReadAll(part)
		require.NoError(t, readErr)
		fields[part.FormName()] = string(data)
	}
	return fields
}
