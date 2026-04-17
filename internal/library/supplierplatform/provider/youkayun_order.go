package supplierprovider

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// SupportsNativeQuantity 表示优卡云支持一次性按 quantity 下单。
func (youkayunProvider) SupportsNativeQuantity() bool { return true }

func (youkayunProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, input CreateOrderInput, baseURL string) (*http.Request, error) {
	fields := map[string]string{
		"userid":     strings.TrimSpace(account.TokenID),
		"goodsid":    strings.TrimSpace(input.SupplierGoodsNo),
		"quantity":   strconv.Itoa(input.Quantity),
		"outorderno": strings.TrimSpace(input.ProviderRequestOrderNo),
	}
	if value := strings.TrimSpace(payloadPrimaryValue(input.Payload)); value != "" {
		fields["accountname"] = value
	}
	if callbackURL, ok := account.ExtraConfig["callback_url"].(string); ok && strings.TrimSpace(callbackURL) != "" {
		fields["callbackurl"] = strings.TrimSpace(callbackURL)
	}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/buygoods", fields, nil)
}

func (youkayunProvider) ParseCreateOrderResponse(statusCode int, body []byte) (*CreateOrderResult, error) {
	raw := string(body)
	if statusCode != http.StatusOK {
		return &CreateOrderResult{
			FinalFailed:   true,
			ErrorCategory: "server_error",
			ErrorMessage:  "http status not ok",
			RawPayload:    raw,
		}, nil
	}
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, err
	}
	code := codeString(payload["code"])
	if code != "1000" {
		message := responseMessage(payload)
		if message == "" {
			message = "下单失败"
		}
		return &CreateOrderResult{
			FinalFailed:   true,
			ErrorCategory: "unknown",
			ErrorCode:     code,
			ErrorMessage:  message,
			RawPayload:    raw,
		}, nil
	}
	data := firstDataMap(payload["data"])
	channelOrderNo := strings.TrimSpace(fmt.Sprint(data["ordersn"]))
	if channelOrderNo == "<nil>" {
		channelOrderNo = ""
	}
	return &CreateOrderResult{
		Accepted:      true,
		ChannelOrderNo: channelOrderNo,
		UpstreamStatus: "accepted",
		RawPayload:     raw,
	}, nil
}

func (youkayunProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, input QueryOrderInput, baseURL string) (*http.Request, error) {
	fields := map[string]string{
		"userid": strings.TrimSpace(account.TokenID),
	}
	if value := strings.TrimSpace(input.ChannelOrderNo); value != "" {
		fields["orderno"] = value
	} else {
		fields["outer_order_id"] = strings.TrimSpace(input.ProviderRequestOrderNo)
	}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/queryorder", fields, nil)
}

func (youkayunProvider) ParseQueryOrderResponse(statusCode int, body []byte) (*QueryOrderResult, error) {
	raw := string(body)
	if statusCode != http.StatusOK {
		return &QueryOrderResult{
			Processing:    true,
			ErrorCategory: "server_error",
			ErrorMessage:  "http status not ok",
			RawPayload:    raw,
		}, nil
	}
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, err
	}
	code := codeString(payload["code"])
	if code != "1000" {
		message := responseMessage(payload)
		if message == "" {
			message = "查单失败"
		}
		return &QueryOrderResult{
			Processing:    true,
			ErrorCategory: "unknown",
			ErrorCode:     code,
			ErrorMessage:  message,
			RawPayload:    raw,
		}, nil
	}

	data := firstDataMap(payload["data"])
	status := codeString(data["status"])
	channelOrderNo := strings.TrimSpace(fmt.Sprint(data["ordersn"]))
	if channelOrderNo == "<nil>" {
		channelOrderNo = ""
	}

	result := &QueryOrderResult{
		ChannelOrderNo: channelOrderNo,
		UpstreamStatus: status,
		RawPayload:     raw,
	}
	switch status {
	case "3":
		result.FinalSuccess = true
	case "5":
		result.FinalFailed = true
	default:
		result.Processing = true
	}
	return result, nil
}

func firstDataMap(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case []any:
		if len(typed) == 0 {
			return map[string]any{}
		}
		if item, ok := typed[0].(map[string]any); ok {
			return item
		}
	}
	return map[string]any{}
}

func payloadPrimaryValue(payload map[string]any) string {
	if payload == nil {
		return ""
	}

	priority := []string{"mobile", "qq", "wechat", "email", "url", "account", "value", "accountname"}
	for _, key := range priority {
		value := strings.TrimSpace(fmt.Sprint(payload[key]))
		if value == "" || value == "<nil>" {
			continue
		}
		return value
	}

	type kv struct {
		Key   string
		Value string
	}
	candidates := make([]kv, 0, len(payload))
	for key, value := range payload {
		v := strings.TrimSpace(fmt.Sprint(value))
		if v == "" || v == "<nil>" {
			continue
		}
		candidates = append(candidates, kv{Key: strings.TrimSpace(key), Value: v})
	}
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0].Value
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Key < candidates[j].Key })
	return candidates[0].Value
}

