package supplierprovider

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func kayixinHeaders(account AccountConfig, now time.Time, body string) map[string]string {
	version := "3.0"
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	return map[string]string{
		"X-App-Id":    strings.TrimSpace(account.TokenID),
		"X-Signature": md5Lower(strings.TrimSpace(account.TokenID) + strings.TrimSpace(account.SecretKey) + version + timestamp + body),
		"X-Timestamp": timestamp,
		"X-Version":   version,
	}
}

func kasushouHeaders(account AccountConfig, now time.Time, body string) map[string]string {
	timestamp := strconv.FormatInt(now.UnixMilli(), 10)
	return map[string]string{
		"Sign":      sha1Lower(timestamp + body + strings.TrimSpace(account.SecretKey)),
		"Timestamp": timestamp,
		"UserId":    strings.TrimSpace(account.TokenID),
	}
}

func xingquanyiBaseParams(account AccountConfig, now time.Time) map[string]string {
	return map[string]string{
		"customer_id": strings.TrimSpace(account.TokenID),
		"timestamp":   strconv.FormatInt(now.Unix(), 10),
	}
}

func julangyunHeaders(account AccountConfig, now time.Time) map[string]string {
	headers := map[string]string{
		"timestamp": strconv.FormatInt(now.UnixMilli(), 10),
		"userCode":  strings.TrimSpace(account.TokenID),
	}
	headers["sign"] = md5Lower(concatSortedNameValues(headers, "sign") + strings.TrimSpace(account.SecretKey))
	return headers
}

func xinghaiSign(params map[string]string, secretKey string) string {
	signParams := make(map[string]string, len(params)+1)
	for key, value := range params {
		signParams[key] = value
	}
	signParams["appSecret"] = strings.TrimSpace(secretKey)
	return md5Lower(sortedQuery(signParams, "sign"))
}

func feisuyuanSign(params map[string]string, secretKey string) string {
	signParams := make(map[string]string, len(params))
	for key, value := range params {
		if key == "sign" || key == "version" || key == "notifyUrl" {
			continue
		}
		signParams[key] = value
	}
	return md5Upper(sortedQuery(signParams) + "&key=" + strings.TrimSpace(secretKey))
}

func stringMapValues(params map[string]string) url.Values {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values
}

func decodeOrderPayload(body []byte) (map[string]any, string, error) {
	raw := string(body)
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, raw, ErrSupplierUnknownResponse
	}
	return payload, raw, nil
}

func mustIntString(value string) int {
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func decimalStringOrZero(value string) string {
	amount, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return "0"
	}
	return amount.Round(4).StringFixed(4)
}

func firstDataItem(payload map[string]any) map[string]any {
	data, ok := payload["data"].([]any)
	if !ok || len(data) == 0 {
		return map[string]any{}
	}
	row, _ := data[0].(map[string]any)
	if row == nil {
		return map[string]any{}
	}
	return row
}

func mapSupplierStatus(value string, successCodes, failedCodes, processingCodes []string) string {
	switch {
	case containsCode(successCodes, value):
		return SupplierOrderStatusSuccess
	case containsCode(failedCodes, value):
		return SupplierOrderStatusFailed
	case containsCode(processingCodes, value):
		return SupplierOrderStatusProcessing
	default:
		return SupplierOrderStatusUnknown
	}
}

func containsCode(codes []string, value string) bool {
	for _, code := range codes {
		if code == value {
			return true
		}
	}
	return false
}

func nonEmptyText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "<nil>" {
		return fallback
	}
	return value
}

func firstNonEmptyText(values ...any) string {
	for _, value := range values {
		text := codeString(value)
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}
