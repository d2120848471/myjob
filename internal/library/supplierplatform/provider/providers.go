package supplierprovider

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type kakayunProvider struct{}
type kayixinProvider struct{}
type kasushouProvider struct{}
type xingquanyiProvider struct{}
type youkayunProvider struct{}
type julangyunProvider struct{}
type feisuyuanProvider struct{}
type xinghaiProvider struct{}

func (kakayunProvider) Code() string { return "kakayun" }
func (kakayunProvider) Name() string { return "卡卡云" }
func (kakayunProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (kakayunProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	timestamp := now.Unix()
	signParams := map[string]string{
		"userid":    strings.TrimSpace(account.TokenID),
		"timestamp": strconv.FormatInt(timestamp, 10),
	}
	payload := struct {
		UserID    string `json:"userid"`
		Timestamp int64  `json:"timestamp"`
		Sign      string `json:"sign"`
	}{
		UserID:    strings.TrimSpace(account.TokenID),
		Timestamp: timestamp,
	}
	payload.Sign = md5Lower(sortedQuery(signParams) + strings.TrimSpace(account.SecretKey))
	return newJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/dockapiv3/user/info", payload, map[string]string{
		"User-Agent": "curl/7.81.0",
	})
}
func (kakayunProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "1", "data", "money")
}

func (kayixinProvider) Code() string { return "kayixin" }
func (kayixinProvider) Name() string { return "卡易信" }
func (kayixinProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (kayixinProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	version := "3.0"
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	bodyString := ""
	if useObject, ok := account.ExtraConfig["sign_with_empty_object"].(bool); ok && useObject {
		bodyString = "{}"
	}
	signValue := md5Lower(strings.TrimSpace(account.TokenID) + strings.TrimSpace(account.SecretKey) + version + timestamp + bodyString)
	headers := map[string]string{
		"X-Version":   version,
		"X-App-Id":    strings.TrimSpace(account.TokenID),
		"X-Timestamp": timestamp,
		"X-Signature": signValue,
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/user/getAccount", bodyString, headers)
}
func (kayixinProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "1000", "data", "balance")
}

func (kasushouProvider) Code() string { return "kasushou" }
func (kasushouProvider) Name() string { return "卡速售" }
func (kasushouProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (kasushouProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	bodyString := "{}"
	headers := map[string]string{
		"UserId":    strings.TrimSpace(account.TokenID),
		"Timestamp": timestamp,
		"Sign":      sha1Lower(timestamp + bodyString + strings.TrimSpace(account.SecretKey)),
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v1/user/info", bodyString, headers)
}
func (kasushouProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "200", "data", "balance")
}

func (xingquanyiProvider) Code() string { return "xingquanyi" }
func (xingquanyiProvider) Name() string { return "星权益" }
func (xingquanyiProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (xingquanyiProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	params := map[string]string{
		"customer_id": strings.TrimSpace(account.TokenID),
		"timestamp":   strconv.FormatInt(now.Unix(), 10),
	}
	params["sign"] = md5Lower(strings.TrimSpace(account.SecretKey) + concatSortedNameValues(params))
	return newJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/customer", params, nil)
}
func (xingquanyiProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "ok", "data", "balance")
}

func (youkayunProvider) Code() string { return "youkayun" }
func (youkayunProvider) Name() string { return "优卡云" }
func (youkayunProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (youkayunProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	fields := map[string]string{
		"userid": strings.TrimSpace(account.TokenID),
	}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/getusermoney", fields, nil)
}
func (youkayunProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "1000", "money")
}

func (julangyunProvider) Code() string { return "julangyun" }
func (julangyunProvider) Name() string { return "聚浪云" }
func (julangyunProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (julangyunProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	headers := map[string]string{
		"userCode":  strings.TrimSpace(account.TokenID),
		"timestamp": strconv.FormatInt(now.UnixMilli(), 10),
	}
	headers["sign"] = md5Lower(concatSortedNameValues(headers, "sign") + strings.TrimSpace(account.SecretKey))
	return newGetRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/recharge/user/amount/detail", headers)
}
func (julangyunProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "200", "data", "amount")
}

func (feisuyuanProvider) Code() string { return "feisuyuan" }
func (feisuyuanProvider) Name() string { return "飞速源" }
func (feisuyuanProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (feisuyuanProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	params := map[string]string{
		"merchantId": strings.TrimSpace(account.TokenID),
		"timeStamp":  strconv.FormatInt(now.Unix(), 10),
		"version":    "1.0",
	}
	signPayload := map[string]string{
		"merchantId": params["merchantId"],
		"timeStamp":  params["timeStamp"],
	}
	params["sign"] = md5Upper(sortedQuery(signPayload) + "&key=" + strings.TrimSpace(account.SecretKey))
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/recharge/info", values, nil)
}
func (feisuyuanProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "0000", "balance")
}

func (xinghaiProvider) Code() string { return "xinghai" }
func (xinghaiProvider) Name() string { return "星海" }
func (xinghaiProvider) CandidateBaseURLs(account AccountConfig) []string {
	return configuredBaseURLs(account.Domain, account.BackupDomain)
}
func (xinghaiProvider) BuildRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	params := map[string]string{
		"appId":     strings.TrimSpace(account.TokenID),
		"timestamp": now.Format("20060102150405000"),
	}
	signParams := map[string]string{
		"appId":     params["appId"],
		"timestamp": params["timestamp"],
		"appSecret": strings.TrimSpace(account.SecretKey),
	}
	params["sign"] = md5Lower(sortedQuery(signParams))
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/account/balance", values, nil)
}
func (xinghaiProvider) ParseBalanceResponse(statusCode int, body []byte) (decimal.Decimal, string, error) {
	payload, err := decodeJSONMap(body)
	if err != nil {
		return decimal.Decimal{}, "响应解析失败", err
	}
	return parseSuccessBalance(payload, "00", "balance")
}
