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
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "http://qqlogin.yxp8.cn/dockapiv3/order/create", req.URL.String())
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), `"userid":"10052"`)
	require.Contains(t, string(body), `"goodsid":"720938"`)
	require.Contains(t, string(body), `"attach":"13088888888"`)
	require.Contains(t, string(body), `"usorderno":"O20260424153000123456-T1"`)
	require.Contains(t, string(body), `"sign":"`)
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
