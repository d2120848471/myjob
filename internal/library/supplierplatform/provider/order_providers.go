package supplierprovider

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (kayixinProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	payload := map[string]any{
		"attach":      []map[string]string{{"name": "充值账号", "value": strings.TrimSpace(input.Account)}},
		"count":       input.Quantity,
		"goodsId":     mustIntString(input.SupplierGoodsNo),
		"notifyUrl":   "",
		"outerNumber": strings.TrimSpace(input.SupplierUSOrderNo),
		"safePrice":   decimalJSONNumberOrZero(safePriceValue(input)),
		"sku":         "",
	}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/order/create", body, kayixinHeaders(account, now, body))
}

func (kayixinProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "1000" {
		message := nonEmptyText(responseMessage(payload), "卡易信下单失败")
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	return CreateOrderResult{
		Accepted:          true,
		Status:            SupplierOrderStatusProcessing,
		SupplierOrderNo:   codeString(data["orderNumber"]),
		SupplierUSOrderNo: codeString(data["outerNumber"]),
		SupplierStatus:    SupplierOrderStatusProcessing,
		Message:           responseMessage(payload),
		Raw:               raw,
	}, nil
}

func (kayixinProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	payload := map[string]any{"orderNumber": strings.TrimSpace(input.SupplierOrderNo), "outerNumber": strings.TrimSpace(input.SupplierUSOrderNo)}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/order/getDetail", body, kayixinHeaders(account, now, body))
}

func (kayixinProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "1000" {
		message := nonEmptyText(responseMessage(payload), "卡易信查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	statusCodeText := codeString(data["status"])
	status := mapSupplierStatus(statusCodeText, []string{"3", "9"}, []string{"4", "5"}, []string{"1", "2", "7", "8"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(data["orderNumber"]), SupplierUSOrderNo: codeString(data["outerNumber"]), SupplierStatus: statusCodeText, Receipt: codeString(data["result"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (kasushouProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	payload := map[string]any{
		"attach":           map[string]string{"recharge_account": strings.TrimSpace(input.Account)},
		"external_orderno": strings.TrimSpace(input.SupplierUSOrderNo),
		"id":               mustIntString(input.SupplierGoodsNo),
		"mark":             "",
		"quantity":         input.Quantity,
		"safe_price":       decimalStringOrZero(safePriceValue(input)),
		"url":              "",
	}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v1/order/buy", body, kasushouHeaders(account, now, body))
}

func (kasushouProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "200" {
		message := nonEmptyText(responseMessage(payload), "卡速售下单失败")
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(data["ordersn"]), SupplierUSOrderNo: codeString(data["external_orderno"]), SupplierStatus: SupplierOrderStatusProcessing, Message: responseMessage(payload), Raw: raw}, nil
}

func (kasushouProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	payload := map[string]any{"day": "0", "external_orderno": strings.TrimSpace(input.SupplierUSOrderNo), "ordersn": strings.TrimSpace(input.SupplierOrderNo)}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v1/order/info", body, kasushouHeaders(account, now, body))
}

func (kasushouProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "200" {
		message := nonEmptyText(responseMessage(payload), "卡速售查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	data := firstDataItem(payload)
	statusCodeText := codeString(data["status"])
	status := mapSupplierStatus(statusCodeText, []string{"3"}, []string{"4", "5", "-1"}, []string{"1", "2"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(data["ordersn"]), SupplierUSOrderNo: codeString(data["external_orderno"]), SupplierStatus: statusCodeText, Receipt: codeString(data["recharge_hints"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (xingquanyiProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	signParams := xingquanyiBaseParams(account, now)
	signParams["notify_url"] = ""
	signParams["outer_order_id"] = strings.TrimSpace(input.SupplierUSOrderNo)
	signParams["product_id"] = strings.TrimSpace(input.SupplierGoodsNo)
	signParams["quantity"] = strconv.Itoa(input.Quantity)
	signParams["recharge_account"] = strings.TrimSpace(input.Account)
	signParams["safe_cost"] = decimalStringOrZero(safePriceValue(input))
	signParams["sign"] = md5Lower(strings.TrimSpace(account.SecretKey) + concatSortedNameValues(signParams))
	payload := map[string]any{
		"customer_id":      signParams["customer_id"],
		"notify_url":       signParams["notify_url"],
		"outer_order_id":   signParams["outer_order_id"],
		"product_id":       mustIntString(input.SupplierGoodsNo),
		"quantity":         input.Quantity,
		"recharge_account": signParams["recharge_account"],
		"safe_cost":        signParams["safe_cost"],
		"sign":             signParams["sign"],
		"timestamp":        signParams["timestamp"],
	}
	return newJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/buy", payload, nil)
}

func (xingquanyiProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "ok" {
		message := nonEmptyText(responseMessage(payload), "星权益下单失败")
		if codeString(payload["code"]) == "server_error" {
			return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusUnknown, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, ErrSupplierUnknownResponse
		}
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(data["order_id"]), SupplierUSOrderNo: codeString(data["outer_order_id"]), SupplierStatus: codeString(data["state"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (xingquanyiProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	params := xingquanyiBaseParams(account, now)
	params["order_id"] = strings.TrimSpace(input.SupplierOrderNo)
	params["outer_order_id"] = strings.TrimSpace(input.SupplierUSOrderNo)
	params["sign"] = md5Lower(strings.TrimSpace(account.SecretKey) + concatSortedNameValues(params))
	return newJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/order", params, nil)
}

func (xingquanyiProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "ok" {
		message := nonEmptyText(responseMessage(payload), "星权益查单失败")
		if codeString(payload["code"]) == "server_error" {
			return QueryOrderResult{Status: SupplierOrderStatusUnknown, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, ErrSupplierUnknownResponse
		}
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	statusCodeText := codeString(data["state"])
	status := mapSupplierStatus(statusCodeText, []string{"200"}, []string{"500"}, []string{"100", "101"})
	if statusCodeText == "501" {
		status = SupplierOrderStatusUnknown
	}
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(data["id"]), SupplierUSOrderNo: codeString(data["outer_order_id"]), SupplierStatus: statusCodeText, Receipt: codeString(data["recharge_info"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (youkayunProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	fields := map[string]string{
		"accountname": strings.TrimSpace(input.Account),
		"callbackurl": "",
		"goodsid":     strings.TrimSpace(input.SupplierGoodsNo),
		"maxmoney":    decimalStringOrZero(safePriceValue(input)),
		"outorderno":  strings.TrimSpace(input.SupplierUSOrderNo),
		"quantity":    strconv.Itoa(input.Quantity),
		"userid":      strings.TrimSpace(account.TokenID),
	}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/buygoods", fields, nil)
}

func (youkayunProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "1000" {
		message := nonEmptyText(responseMessage(payload), "优卡云下单失败")
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(data["ordersn"]), SupplierUSOrderNo: codeString(data["outorderno"]), SupplierStatus: SupplierOrderStatusProcessing, Message: responseMessage(payload), Raw: raw}, nil
}

func (youkayunProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	fields := map[string]string{"orderno": strings.TrimSpace(input.SupplierOrderNo), "outer_order_id": strings.TrimSpace(input.SupplierUSOrderNo), "userid": strings.TrimSpace(account.TokenID)}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/queryorder", fields, nil)
}

func (youkayunProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "1000" {
		message := nonEmptyText(responseMessage(payload), "优卡云查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	statusCodeText := codeString(data["status"])
	status := mapSupplierStatus(statusCodeText, []string{"3"}, []string{"5"}, []string{"1", "2"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(data["ordersn"]), SupplierUSOrderNo: firstNonEmptyText(data["outer_order_no"], data["outorderno"]), SupplierStatus: statusCodeText, Receipt: codeString(data["result"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (julangyunProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	payload := map[string]any{
		"accessOrderNo":   strings.TrimSpace(input.SupplierUSOrderNo),
		"accessPrice":     decimalJSONNumberOrZero(safePriceValue(input)),
		"callbackUrl":     "",
		"goodsCode":       strings.TrimSpace(input.SupplierGoodsNo),
		"orderNum":        input.Quantity,
		"rechargeAccount": strings.TrimSpace(input.Account),
	}
	return newSortedJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/recharge/order/submit", payload, julangyunHeaders(account, now))
}

func (julangyunProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "200" {
		message := nonEmptyText(responseMessage(payload), "聚浪云下单失败")
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: codeString(payload["code"]), Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(data["returnOrderNo"]), SupplierUSOrderNo: codeString(data["accessOrderNo"]), SupplierStatus: codeString(data["orderStatus"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (julangyunProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	payload := map[string]any{"accessOrderNo": strings.TrimSpace(input.SupplierUSOrderNo), "returnOrderNo": strings.TrimSpace(input.SupplierOrderNo)}
	return newSortedJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/recharge/order/detail", payload, julangyunHeaders(account, now))
}

func (julangyunProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "200" {
		message := nonEmptyText(responseMessage(payload), "聚浪云查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	data := nestedMap(payload, "data")
	statusCodeText := codeString(data["orderStatus"])
	status := mapSupplierStatus(statusCodeText, []string{"30"}, []string{"40", "50"}, []string{"20"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(data["returnOrderNo"]), SupplierUSOrderNo: codeString(data["accessOrderNo"]), SupplierStatus: statusCodeText, Receipt: codeString(data["remark"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (xinghaiProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	params := map[string]string{
		"amount":      strconv.Itoa(input.Quantity),
		"appId":       strings.TrimSpace(account.TokenID),
		"callbackUrl": "",
		"itemId":      strings.TrimSpace(input.SupplierGoodsNo),
		"itemPrice":   formatXinghaiItemPrice(safePriceValue(input)),
		"outOrderId":  strings.TrimSpace(input.SupplierUSOrderNo),
		"timestamp":   now.Format("20060102150405000"),
		"uuid":        strings.TrimSpace(input.Account),
	}
	params["sign"] = xinghaiSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/order/submit", stringMapValues(params), nil)
}

func (xinghaiProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	code := codeString(payload["code"])
	if code != "00" {
		message := nonEmptyText(responseMessage(payload), "星海下单失败")
		if code == "-16" || code == "-99" {
			return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusUnknown, SupplierStatus: code, Message: message, Raw: raw}, ErrSupplierUnknownResponse
		}
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: code, Message: message, Raw: raw}, errors.New(message)
	}
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(payload["orderId"]), SupplierUSOrderNo: codeString(payload["outOrderId"]), SupplierStatus: code, Message: responseMessage(payload), Raw: raw}, nil
}

func (xinghaiProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	params := map[string]string{"appId": strings.TrimSpace(account.TokenID), "orderId": strings.TrimSpace(input.SupplierOrderNo), "outOrderId": strings.TrimSpace(input.SupplierUSOrderNo), "timestamp": now.Format("20060102150405000")}
	params["sign"] = xinghaiSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/order/query", stringMapValues(params), nil)
}

func (xinghaiProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "00" {
		message := nonEmptyText(responseMessage(payload), "星海查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	statusCodeText := codeString(payload["orderStatus"])
	status := mapSupplierStatus(statusCodeText, []string{"2"}, []string{"3"}, []string{"0", "1"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(payload["orderId"]), SupplierUSOrderNo: codeString(payload["outOrderId"]), SupplierStatus: statusCodeText, Receipt: codeString(payload["orderDesc"]), Message: responseMessage(payload), Raw: raw}, nil
}

func (feisuyuanProvider) BuildCreateOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input CreateOrderInput) (*http.Request, error) {
	params := map[string]string{
		"accountType":     feisuyuanAccountType(account.ExtraConfig),
		"merchantId":      strings.TrimSpace(account.TokenID),
		"notifyUrl":       "",
		"number":          "1",
		"outTradeNo":      strings.TrimSpace(input.SupplierUSOrderNo),
		"productId":       strings.TrimSpace(input.SupplierGoodsNo),
		"rechargeAccount": strings.TrimSpace(input.Account),
		"timeStamp":       strconv.FormatInt(now.Unix(), 10),
		"version":         "1.0",
	}
	params["sign"] = feisuyuanSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/recharge/order", stringMapValues(params), nil)
}

func (feisuyuanProvider) ParseCreateOrderResponse(statusCode int, body []byte) (CreateOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return CreateOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	code := codeString(payload["code"])
	if code != "2000" {
		message := nonEmptyText(responseMessage(payload), "飞速源下单失败")
		if code == "1999" {
			return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusUnknown, SupplierStatus: code, Message: message, Raw: raw}, ErrSupplierUnknownResponse
		}
		return CreateOrderResult{Accepted: false, Status: SupplierOrderStatusFailed, SupplierStatus: code, Message: message, Raw: raw}, errors.New(message)
	}
	return CreateOrderResult{Accepted: true, Status: SupplierOrderStatusProcessing, SupplierOrderNo: codeString(payload["orderNo"]), SupplierUSOrderNo: codeString(payload["outTradeNo"]), SupplierStatus: code, Message: responseMessage(payload), Raw: raw}, nil
}

func (feisuyuanProvider) BuildQueryOrderRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input QueryOrderInput) (*http.Request, error) {
	params := map[string]string{"merchantId": strings.TrimSpace(account.TokenID), "outTradeNo": strings.TrimSpace(input.SupplierUSOrderNo), "timeStamp": strconv.FormatInt(now.Unix(), 10), "version": "1.0"}
	params["sign"] = feisuyuanSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/recharge/query", stringMapValues(params), nil)
}

func (feisuyuanProvider) ParseQueryOrderResponse(statusCode int, body []byte) (QueryOrderResult, error) {
	payload, raw, err := decodeOrderPayload(body)
	if err != nil {
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Raw: raw}, err
	}
	if codeString(payload["code"]) != "0000" {
		message := nonEmptyText(responseMessage(payload), "飞速源查单失败")
		return QueryOrderResult{Status: SupplierOrderStatusUnknown, Message: message, Raw: raw}, errors.New(message)
	}
	statusCodeText := codeString(payload["status"])
	status := mapSupplierStatus(statusCodeText, []string{"01"}, []string{"03"}, []string{"02"})
	return QueryOrderResult{Status: status, SupplierOrderNo: codeString(payload["orderNo"]), SupplierUSOrderNo: codeString(payload["outTradeNo"]), SupplierStatus: statusCodeText, Receipt: responseMessage(payload), Message: responseMessage(payload), Raw: raw}, nil
}
