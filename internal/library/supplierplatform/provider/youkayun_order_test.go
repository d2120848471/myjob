package supplierprovider

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYoukayunOrderProvider_BuildCreateOrderRequest(t *testing.T) {
	provider := youkayunProvider{}
	account := AccountConfig{
		TokenID:     "merchant001",
		SecretKey:   "secretXYZ",
		ExtraConfig: map[string]any{},
	}
	input := CreateOrderInput{
		ProviderRequestOrderNo: "PR001",
		SupplierGoodsNo:        "100",
		Quantity:               2,
		Payload: map[string]any{
			"mobile": "13800138000",
		},
	}

	req, err := provider.BuildCreateOrderRequest(context.Background(), account, input, "http://platform.example.com")
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.Method)
	require.Equal(t, "http://platform.example.com/api/buygoods", req.URL.String())

	body := readRequestBody(t, req)
	fields := decodeMultipartFields(t, req, body)
	require.Equal(t, "merchant001", fields["userid"])
	require.Equal(t, "100", fields["goodsid"])
	require.Equal(t, "2", fields["quantity"])
	require.Equal(t, "PR001", fields["outorderno"])
	require.Equal(t, "13800138000", fields["accountname"])

	expectedSign := md5Lower(sortedQuery(map[string]string{
		"userid":     fields["userid"],
		"goodsid":    fields["goodsid"],
		"quantity":   fields["quantity"],
		"outorderno": fields["outorderno"],
		"accountname": fields["accountname"],
	}) + "secretXYZ")
	require.Equal(t, expectedSign, fields["sign"])
}

func TestYoukayunOrderProvider_ParseCreateOrderResponse(t *testing.T) {
	provider := youkayunProvider{}

	success, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH001"}}`))
	require.NoError(t, err)
	require.True(t, success.Accepted)
	require.Equal(t, "CH001", success.ChannelOrderNo)

	failed, err := provider.ParseCreateOrderResponse(http.StatusOK, []byte(`{"code":1002,"msg":"签名错误"}`))
	require.NoError(t, err)
	require.True(t, failed.FinalFailed)
	require.Equal(t, "1002", failed.ErrorCode)
	require.Equal(t, "签名错误", failed.ErrorMessage)
}

func TestYoukayunOrderProvider_BuildQueryOrderRequest_UsesChannelOrderNoWhenPresent(t *testing.T) {
	provider := youkayunProvider{}
	account := AccountConfig{
		TokenID:     "merchant001",
		SecretKey:   "secretXYZ",
		ExtraConfig: map[string]any{},
	}
	input := QueryOrderInput{
		ProviderRequestOrderNo: "PR001",
		ChannelOrderNo:         "CH001",
	}

	req, err := provider.BuildQueryOrderRequest(context.Background(), account, input, "http://platform.example.com")
	require.NoError(t, err)
	require.Equal(t, "http://platform.example.com/api/queryorder", req.URL.String())

	body := readRequestBody(t, req)
	fields := decodeMultipartFields(t, req, body)
	require.Equal(t, "merchant001", fields["userid"])
	require.Equal(t, "CH001", fields["orderno"])
	require.Empty(t, fields["outer_order_id"])
}

func TestYoukayunOrderProvider_ParseQueryOrderResponse_StatusMapping(t *testing.T) {
	provider := youkayunProvider{}

	processing, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH001","status":1}}`))
	require.NoError(t, err)
	require.True(t, processing.Processing)

	success, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH001","status":3}}`))
	require.NoError(t, err)
	require.True(t, success.FinalSuccess)

	failed, err := provider.ParseQueryOrderResponse(http.StatusOK, []byte(`{"code":1000,"msg":"ok","data":{"ordersn":"CH001","status":5}}`))
	require.NoError(t, err)
	require.True(t, failed.FinalFailed)
}

