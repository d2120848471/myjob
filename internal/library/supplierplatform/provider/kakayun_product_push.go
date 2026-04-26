package supplierprovider

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const kakayunProductPushBaseURL = "http://public.kky.v3.api.kakayun.vip"
const kakayunPushTimestampSkew = 5 * time.Minute

func (kakayunProvider) BuildSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error) {
	return kakayunProductSubscribeRequest(ctx, account, now, input, "/dockapiv3/goods/subscribe")
}

func (kakayunProvider) BuildCancelSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput) (*http.Request, error) {
	return kakayunProductSubscribeRequest(ctx, account, now, input, "/dockapiv3/goods/cancelsubscribe")
}

func kakayunProductSubscribeRequest(ctx context.Context, account AccountConfig, now time.Time, input ProductSubscribeInput, path string) (*http.Request, error) {
	payload := map[string]any{
		"userid":    strings.TrimSpace(account.TokenID),
		"timestamp": now.Unix(),
		"goodsid":   strings.TrimSpace(input.SupplierGoodsNo),
	}
	payload["sign"] = kakayunSign(payload, account.SecretKey)
	return newJSONRequest(ctx, kakayunProductPushBaseURL+path, payload, map[string]string{"User-Agent": "curl/7.81.0"})
}

func (kakayunProvider) ParseMutationResponse(statusCode int, body []byte) (string, error) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return "", errors.New("卡卡云订阅接口 HTTP 状态异常: " + strconv.Itoa(statusCode))
	}
	payload, err := decodeJSONMap(body)
	if err != nil {
		return "", err
	}
	message := responseMessage(payload)
	if codeString(payload["code"]) != "1" {
		if message == "" {
			message = "卡卡云订阅接口失败"
		}
		return message, errors.New(message)
	}
	if message == "" {
		message = "成功"
	}
	return message, nil
}

func (kakayunProvider) ParseProductChangePush(account AccountConfig, now time.Time, body []byte) (ProductChangePushResult, error) {
	raw := string(body)
	payload, err := decodeJSONMap(body)
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	timestamp, err := int64FromRequired(payload["timestamp"])
	if err != nil {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送缺少有效时间戳")
	}
	if delta := now.Sub(time.Unix(timestamp, 0)); delta > kakayunPushTimestampSkew || delta < -kakayunPushTimestampSkew {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送时间戳已过期")
	}
	expected := kakayunSign(payload, account.SecretKey)
	actual := strings.TrimSpace(codeString(payload["sign"]))
	if !kakayunSignEqual(expected, actual) {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送签名错误")
	}
	supplierGoodsNo := codeString(payload["goodsid"])
	if supplierGoodsNo == "" {
		return ProductChangePushResult{Raw: raw}, errors.New("卡卡云推送缺少商品编号")
	}
	price, err := decimalFromValue(payload["goodsprice"])
	priceValid := err == nil && !price.IsNegative()
	if priceValid {
		price = price.Round(4)
	}
	return ProductChangePushResult{
		SupplierGoodsNo: supplierGoodsNo,
		GoodsName:       strings.TrimSpace(codeString(payload["goodsname"])),
		GoodsPrice:      price,
		GoodsPriceValid: priceValid,
		GoodsStatus:     codeString(payload["goodsstatus"]),
		Raw:             raw,
	}, nil
}

func kakayunSignEqual(expected, actual string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	actual = strings.ToLower(strings.TrimSpace(actual))
	if len(expected) != len(actual) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

func int64FromRequired(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case float64:
		return int64(typed), nil
	case json.Number:
		return typed.Int64()
	case string:
		return strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
	default:
		return 0, fmt.Errorf("字段不是有效整数")
	}
}
