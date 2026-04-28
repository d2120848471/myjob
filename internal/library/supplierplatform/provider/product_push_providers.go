package supplierprovider

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

func (kasushouProvider) ParseProductChangePush(account AccountConfig, now time.Time, body []byte) (ProductChangePushResult, error) {
	payload, raw, err := productChangePushPayload(body)
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	timestamp, err := int64FromRequired(payload["time"])
	if err != nil {
		return ProductChangePushResult{Raw: raw}, errors.New("卡速售推送缺少有效时间戳")
	}
	if delta := now.Sub(time.UnixMilli(timestamp)); delta > kakayunPushTimestampSkew || delta < -kakayunPushTimestampSkew {
		return ProductChangePushResult{Raw: raw}, errors.New("卡速售推送时间戳已过期")
	}
	signBody, err := orderedJSONBody(map[string]any{"id": codeString(payload["id"]), "time": codeString(payload["time"])})
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	expected := sha1Lower(codeString(payload["time"]) + signBody + strings.TrimSpace(account.SecretKey))
	if !kakayunSignEqual(expected, codeString(payload["sign"])) {
		return ProductChangePushResult{Raw: raw}, errors.New("卡速售推送签名错误")
	}
	return productChangePushResultFromFields(raw, codeString(payload["id"]), codeString(payload["goods_name"]), payload["goods_price"], payload["status"], "卡速售推送缺少商品编号")
}

func (xingquanyiProvider) ParseProductChangePush(account AccountConfig, _ time.Time, body []byte) (ProductChangePushResult, error) {
	payload, raw, err := productChangePushPayload(body)
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	expected := xingquanyiPushSign(payload, account.SecretKey)
	if !kakayunSignEqual(expected, codeString(payload["sign"])) {
		return ProductChangePushResult{Raw: raw}, errors.New("星权益推送签名错误")
	}
	eventData := productChangeEventData(payload["event_data"])
	priceValue := payload["goods_price"]
	if _, ok := payload["goods_price"]; !ok {
		priceValue = eventData["price"]
	}
	name := codeString(payload["product_name"])
	if name == "" {
		name = codeString(eventData["product_name"])
	}
	status := payload["status"]
	if _, ok := payload["status"]; !ok {
		status = firstNonEmptyValue(eventData["status"], eventData["supply_state"], eventData["hold_state"], eventData["stock_state"])
	}
	return productChangePushResultFromFields(raw, firstNonEmptyText(payload["product_id"], payload["id"]), name, priceValue, status, "星权益推送缺少商品编号")
}

func (youkayunProvider) ParseProductChangePush(account AccountConfig, _ time.Time, body []byte) (ProductChangePushResult, error) {
	payload, raw, err := productChangePushPayload(body)
	if err != nil {
		return ProductChangePushResult{Raw: raw}, err
	}
	expected := youkayunPushSign(payload, account.SecretKey)
	if !kakayunSignEqual(expected, codeString(payload["sign"])) {
		return ProductChangePushResult{Raw: raw}, errors.New("优卡云推送签名错误")
	}
	return productChangePushResultFromFields(raw, firstNonEmptyText(payload["goods_id"], payload["goodsid"], payload["id"]), firstNonEmptyText(payload["goods_name"], payload["name"]), firstNonEmptyValue(payload["goods_price"], payload["price"]), firstNonEmptyValue(payload["status"], payload["goods_status"]), "优卡云推送缺少商品编号")
}

func productChangePushPayload(body []byte) (map[string]any, string, error) {
	raw := string(body)
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, raw, err
	}
	return payload, raw, nil
}

func productChangeEventData(value any) map[string]any {
	if payload, ok := value.(map[string]any); ok {
		return payload
	}
	text := strings.TrimSpace(codeString(value))
	if text == "" || text == "<nil>" {
		return map[string]any{}
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

func productChangePushResultFromFields(raw, supplierGoodsNo, goodsName string, priceValue any, statusValue any, missingMessage string) (ProductChangePushResult, error) {
	supplierGoodsNo = strings.TrimSpace(supplierGoodsNo)
	if supplierGoodsNo == "" {
		return ProductChangePushResult{Raw: raw}, errors.New(missingMessage)
	}
	price, err := decimalFromValue(priceValue)
	priceValid := err == nil && !price.IsNegative()
	if priceValid {
		price = price.Round(4)
	}
	return ProductChangePushResult{
		SupplierGoodsNo: supplierGoodsNo,
		GoodsName:       strings.TrimSpace(goodsName),
		GoodsPrice:      price,
		GoodsPriceValid: priceValid,
		GoodsStatus:     codeString(statusValue),
		Raw:             raw,
	}, nil
}

func xingquanyiPushSign(payload map[string]any, secretKey string) string {
	params := make(map[string]string, len(payload))
	for key, value := range payload {
		if key == "sign" || value == nil {
			continue
		}
		text := pushSignValue(value)
		if strings.TrimSpace(text) == "" {
			continue
		}
		params[key] = text
	}
	return md5Lower(strings.TrimSpace(secretKey) + concatSortedNameValues(params))
}

func youkayunPushSign(payload map[string]any, secretKey string) string {
	params := make(map[string]string, len(payload))
	for key, value := range payload {
		if key == "sign" || value == nil {
			continue
		}
		text := pushSignValue(value)
		if strings.TrimSpace(text) == "" {
			continue
		}
		params[key] = text
	}
	return md5Lower(sortedQuery(params) + strings.TrimSpace(secretKey))
}

func pushSignValue(value any) string {
	switch value.(type) {
	case map[string]any, []any:
		raw, err := json.Marshal(value)
		if err != nil {
			return ""
		}
		return string(raw)
	default:
		return codeString(value)
	}
}

func firstNonEmptyValue(values ...any) any {
	for _, value := range values {
		if strings.TrimSpace(codeString(value)) != "" && strings.TrimSpace(codeString(value)) != "<nil>" {
			return value
		}
	}
	return nil
}
