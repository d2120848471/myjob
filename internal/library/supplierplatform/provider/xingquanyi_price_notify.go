package supplierprovider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func (xingquanyiProvider) VerifyPriceNotifySignature(account AccountConfig, headers http.Header, body []byte) error {
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	sign := strings.TrimSpace(fmt.Sprint(payload["sign"]))
	if sign == "" || sign == "<nil>" {
		return fmt.Errorf("missing sign")
	}

	params := make(map[string]string, len(payload))
	for key, value := range payload {
		if strings.TrimSpace(key) == "" || key == "sign" {
			continue
		}
		params[key] = stringifyXingquanyiValue(value)
	}
	expected := md5Lower(strings.TrimSpace(account.SecretKey) + concatSortedNameValues(params))
	if strings.TrimSpace(strings.ToLower(expected)) != strings.TrimSpace(strings.ToLower(sign)) {
		return fmt.Errorf("invalid sign")
	}
	return nil
}

func (xingquanyiProvider) ParsePriceNotifyPayload(account AccountConfig, headers http.Header, body []byte) (*PriceNotifyResult, error) {
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	customerID := strings.TrimSpace(fmt.Sprint(payload["customer_id"]))
	if customerID == "<nil>" {
		customerID = ""
	}
	productID := strings.TrimSpace(fmt.Sprint(payload["product_id"]))
	if productID == "<nil>" {
		productID = ""
	}
	productName := strings.TrimSpace(fmt.Sprint(payload["product_name"]))
	if productName == "<nil>" {
		productName = ""
	}
	eventType := strings.TrimSpace(fmt.Sprint(payload["event_type"]))
	if eventType == "<nil>" {
		eventType = ""
	}

	timestamp, _ := int64FromAny(payload["timestamp"])
	eventDataRaw := strings.TrimSpace(fmt.Sprint(payload["event_data"]))
	if eventDataRaw == "<nil>" {
		eventDataRaw = ""
	}
	eventData := make(map[string]any)
	if strings.TrimSpace(eventDataRaw) != "" {
		_ = json.Unmarshal([]byte(eventDataRaw), &eventData)
	}

	priceRaw := strings.TrimSpace(fmt.Sprint(eventData["price"]))
	if priceRaw == "<nil>" {
		priceRaw = ""
	}
	price, err := decimal.NewFromString(priceRaw)
	if err != nil {
		price = decimal.Zero
	}

	sign := strings.TrimSpace(fmt.Sprint(payload["sign"]))
	if sign == "<nil>" {
		sign = ""
	}
	idempotencyKey := strings.TrimSpace(sign)
	if idempotencyKey == "" {
		idempotencyKey = strings.TrimSpace(customerID + ":" + productID + ":" + strconv.FormatInt(timestamp, 10) + ":" + eventType)
	}

	return &PriceNotifyResult{
		PlatformAccountLocator: customerID,
		SupplierGoodsNo:        productID,
		SupplierGoodsName:      productName,
		SourceCostPrice:        price,
		NotifyAt:               time.Unix(timestamp, 0),
		IdempotencyKey:          idempotencyKey,
		RawPayload:              string(body),
	}, nil
}

func stringifyXingquanyiValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return strings.TrimSpace(typed.String())
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strings.TrimSpace(strconv.FormatFloat(typed, 'f', -1, 64))
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	default:
		v := strings.TrimSpace(fmt.Sprint(value))
		if v == "<nil>" {
			return ""
		}
		return v
	}
}

func int64FromAny(value any) (int64, bool) {
	switch typed := value.(type) {
	case json.Number:
		v, err := typed.Int64()
		return v, err == nil
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case string:
		v, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return v, err == nil
	default:
		return 0, false
	}
}

