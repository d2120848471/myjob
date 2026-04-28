package supplierprovider

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (kayixinProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	payload := map[string]any{"goodsId": mustIntString(input.SupplierGoodsNo)}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v3/goods/getDetail", body, kayixinHeaders(account, now, body))
}

func (kayixinProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "1000", "卡易信商品详情查询失败")
	if err != nil {
		return ProductInfoResult{Raw: raw}, err
	}
	data := nestedMap(payload, "data")
	return productInfoFromFields(raw, data["goodsId"], data["name"], data["salesPrice"])
}

func (kasushouProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	payload := map[string]any{"id": mustIntString(input.SupplierGoodsNo)}
	body, err := orderedJSONBody(payload)
	if err != nil {
		return nil, err
	}
	return newEmptyJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/v1/goods/info", body, kasushouHeaders(account, now, body))
}

func (kasushouProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "200", "卡速售商品详情查询失败")
	if err != nil {
		return ProductInfoResult{Raw: raw}, err
	}
	data := nestedMap(payload, "data")
	return productInfoFromFields(raw, data["id"], data["goods_name"], data["goods_price"])
}

func (xingquanyiProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	params := xingquanyiBaseParams(account, now)
	params["product_id"] = strings.TrimSpace(input.SupplierGoodsNo)
	params["sign"] = md5Lower(strings.TrimSpace(account.SecretKey) + concatSortedNameValues(params))
	return newJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/product", params, nil)
}

func (xingquanyiProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "ok", "星权益商品详情查询失败")
	if err != nil {
		return ProductInfoResult{Raw: raw}, err
	}
	data := nestedMap(payload, "data")
	name := strings.TrimSpace(codeString(data["product_name"]))
	if alias := strings.TrimSpace(codeString(data["name"])); alias != "" && alias != name {
		name = strings.TrimSpace(name + " " + alias)
	}
	return productInfoFromFields(raw, data["id"], name, data["price"])
}

func (youkayunProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	fields := map[string]string{
		"goodsid": strings.TrimSpace(input.SupplierGoodsNo),
		"userid":  strings.TrimSpace(account.TokenID),
	}
	fields["sign"] = md5Lower(sortedQuery(fields) + strings.TrimSpace(account.SecretKey))
	return newMultipartRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/goodsdetails", fields, nil)
}

func (youkayunProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "1000", "优卡云商品详情查询失败")
	if err != nil {
		return ProductInfoResult{Raw: raw}, err
	}
	data := nestedMap(payload, "data")
	return productInfoFromFields(raw, data["id"], data["goods_name"], data["goods_price"])
}

func (julangyunProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	payload := map[string]any{"goodsCode": strings.TrimSpace(input.SupplierGoodsNo)}
	return newSortedJSONRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/recharge/goods/detail", payload, julangyunHeaders(account, now))
}

func (julangyunProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "200", "聚浪云商品详情查询失败")
	if err != nil {
		return ProductInfoResult{Raw: raw}, err
	}
	data := nestedMap(payload, "data")
	return productInfoFromFields(raw, data["goodsCode"], data["goodsName"], data["goodsPrice"])
}

func (xinghaiProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	return (xinghaiProvider{}).BuildProductInfoListRequest(ctx, account, now, baseURL)
}

func (xinghaiProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	return ProductInfoResult{Raw: string(body)}, errors.New("星海商品信息需要通过列表匹配")
}

func (xinghaiProvider) BuildProductInfoListRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	params := map[string]string{
		"appId":     strings.TrimSpace(account.TokenID),
		"timestamp": now.Format("20060102150405000"),
	}
	params["sign"] = xinghaiSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/api/item/query", stringMapValues(params), nil)
}

func (xinghaiProvider) ParseProductInfoListResponse(statusCode int, body []byte) (map[string]ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "00", "星海商品列表查询失败")
	if err != nil {
		return nil, err
	}
	rows, _ := payload["data"].([]any)
	return productInfoListFromRows(raw, rows, "itemId", "itemName", "price")
}

func (feisuyuanProvider) BuildProductInfoRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string, input ProductInfoInput) (*http.Request, error) {
	return (feisuyuanProvider{}).BuildProductInfoListRequest(ctx, account, now, baseURL)
}

func (feisuyuanProvider) ParseProductInfoResponse(statusCode int, body []byte) (ProductInfoResult, error) {
	return ProductInfoResult{Raw: string(body)}, errors.New("飞速源商品信息需要通过列表匹配")
}

func (feisuyuanProvider) BuildProductInfoListRequest(ctx context.Context, account AccountConfig, now time.Time, baseURL string) (*http.Request, error) {
	params := map[string]string{
		"merchantId": strings.TrimSpace(account.TokenID),
		"timeStamp":  strconv.FormatInt(now.Unix(), 10),
		"version":    "1.0",
	}
	params["sign"] = feisuyuanSign(params, account.SecretKey)
	return newFormRequest(ctx, strings.TrimRight(baseURL, "/")+"/recharge/product", stringMapValues(params), nil)
}

func (feisuyuanProvider) ParseProductInfoListResponse(statusCode int, body []byte) (map[string]ProductInfoResult, error) {
	payload, raw, err := productInfoPayload(body, "0000", "飞速源商品列表查询失败")
	if err != nil {
		return nil, err
	}
	rows, _ := payload["products"].([]any)
	return productInfoListFromRows(raw, rows, "product_id", "item_name", "channel_price")
}

func productInfoPayload(body []byte, expectedCode, fallback string) (map[string]any, string, error) {
	raw := string(body)
	payload, err := decodeJSONMap(body)
	if err != nil {
		return nil, raw, err
	}
	if codeString(payload["code"]) != expectedCode {
		return nil, raw, errors.New(nonEmptyText(responseMessage(payload), fallback))
	}
	return payload, raw, nil
}

func productInfoFromFields(raw string, idValue, nameValue, priceValue any) (ProductInfoResult, error) {
	name := strings.TrimSpace(codeString(nameValue))
	price, err := decimalFromValue(priceValue)
	priceValid := err == nil && !price.IsNegative()
	if !priceValid && name == "" {
		return ProductInfoResult{Raw: raw}, errors.New("供应商商品详情缺少可同步字段")
	}
	if priceValid {
		price = price.Round(4)
	}
	return ProductInfoResult{SupplierGoodsNo: codeString(idValue), GoodsName: name, GoodsPrice: price, GoodsPriceValid: priceValid, Raw: raw}, nil
}

func productInfoListFromRows(raw string, rows []any, idKey, nameKey, priceKey string) (map[string]ProductInfoResult, error) {
	result := make(map[string]ProductInfoResult, len(rows))
	for _, rowValue := range rows {
		row, ok := rowValue.(map[string]any)
		if !ok {
			continue
		}
		info, err := productInfoFromFields(raw, row[idKey], row[nameKey], row[priceKey])
		if err != nil {
			continue
		}
		if info.SupplierGoodsNo != "" {
			result[info.SupplierGoodsNo] = info
		}
	}
	if len(result) == 0 {
		return nil, errors.New("供应商商品列表缺少可同步字段")
	}
	return result, nil
}
