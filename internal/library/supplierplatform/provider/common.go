package supplierprovider

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	return strings.TrimSuffix(host, "/")
}

func configuredBaseURLs(domain, backupDomain string) []string {
	hosts := []string{normalizeHost(domain), normalizeHost(backupDomain)}
	seen := map[string]struct{}{}
	result := make([]string, 0, 4)
	for _, host := range hosts {
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		// HTTP 优先：seller 域名通常不支持 HTTPS API，先尝试 HTTP 可避免 HTTPS
		// 失败后 CDN/WAF 频控导致后续 HTTP 请求被断连（EOF）的问题。
		result = append(result, "http://"+host, "https://"+host)
	}
	return result
}

func sortedKeys(params map[string]string, exclude ...string) []string {
	ignored := make(map[string]struct{}, len(exclude))
	for _, key := range exclude {
		ignored[strings.TrimSpace(key)] = struct{}{}
	}
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if _, ok := ignored[key]; ok {
			continue
		}
		if strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedQuery(params map[string]string, exclude ...string) string {
	keys := sortedKeys(params, exclude...)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	return strings.Join(parts, "&")
}

func concatSortedNameValues(params map[string]string, exclude ...string) string {
	keys := sortedKeys(params, exclude...)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(params[key])
	}
	return builder.String()
}

func md5Lower(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func md5Upper(value string) string {
	return strings.ToUpper(md5Lower(value))
}

func sha1Lower(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func kakayunSign(payload map[string]any, secretKey string) string {
	params := make(map[string]string, len(payload))
	for key, value := range payload {
		if key == "sign" || value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text == "" || text == "<nil>" {
			continue
		}
		params[key] = text
	}
	return md5Lower(sortedQuery(params) + strings.TrimSpace(secretKey))
}

func newJSONRequest(ctx context.Context, url string, payload any, headers map[string]string) (*http.Request, error) {
	body := []byte{}
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func newEmptyJSONRequest(ctx context.Context, url string, body string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func newFormRequest(ctx context.Context, url string, values url.Values, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func newMultipartRequest(ctx context.Context, requestURL string, fields map[string]string, headers map[string]string) (*http.Request, error) {
	buffer := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(buffer)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(buffer.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func newGetRequest(ctx context.Context, requestURL string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func decodeJSONMap(body []byte) (map[string]any, error) {
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func responseMessage(payload map[string]any) string {
	if value := strings.TrimSpace(fmt.Sprint(payload["message"])); value != "" && value != "<nil>" {
		return value
	}
	if value := strings.TrimSpace(fmt.Sprint(payload["msg"])); value != "" && value != "<nil>" {
		return value
	}
	return ""
}

func codeString(value any) string {
	return strings.TrimSpace(fmt.Sprint(value))
}

func nestedMap(payload map[string]any, key string) map[string]any {
	if value, ok := payload[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}

func decimalFromValue(value any) (decimal.Decimal, error) {
	switch typed := value.(type) {
	case string:
		return decimal.NewFromString(strings.TrimSpace(typed))
	case json.Number:
		return decimal.NewFromString(typed.String())
	case float64:
		return decimal.NewFromString(strings.TrimSpace(fmt.Sprintf("%.4f", typed)))
	case float32:
		return decimal.NewFromString(strings.TrimSpace(fmt.Sprintf("%.4f", typed)))
	case int:
		return decimal.NewFromInt(int64(typed)), nil
	case int64:
		return decimal.NewFromInt(typed), nil
	case int32:
		return decimal.NewFromInt32(typed), nil
	default:
		return decimal.Decimal{}, fmt.Errorf("余额字段格式错误")
	}
}

func parseSuccessBalance(payload map[string]any, expectedCode string, valuePath ...string) (decimal.Decimal, string, error) {
	if codeString(payload["code"]) != expectedCode {
		message := responseMessage(payload)
		if message == "" {
			message = "余额查询失败"
		}
		return decimal.Decimal{}, message, errors.New(message)
	}
	value := any(nil)
	current := payload
	for index, key := range valuePath {
		if index == len(valuePath)-1 {
			value = current[key]
			break
		}
		current = nestedMap(current, key)
	}
	amount, err := decimalFromValue(value)
	if err != nil {
		return decimal.Decimal{}, "余额解析失败", err
	}
	message := responseMessage(payload)
	return amount, message, nil
}
