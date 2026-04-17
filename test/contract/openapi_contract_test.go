package contract_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAPIAndSwaggerExposed(t *testing.T) {
	h := newTestHarness(t)

	oaiRes := h.rawRequest("GET", "/api.json", nil, "")
	require.Equal(t, 200, oaiRes.status)
	require.Contains(t, oaiRes.body, "/api/admin/auth/login")
	require.Contains(t, oaiRes.body, "/api/open/orders")
	require.Contains(t, oaiRes.body, "/api/open/orders/{order_no}")
	require.Contains(t, oaiRes.body, "/api/open/orders/by-client/{client_order_no}")
	require.Contains(t, oaiRes.body, "/api/provider/{provider_code}/order-callback")
	require.Contains(t, oaiRes.body, "/api/provider/{provider_code}/price-notify")
	require.Contains(t, oaiRes.body, "BearerAuth")
	require.Contains(t, oaiRes.body, "message")

	swaggerRes := h.rawRequest("GET", "/swagger/", nil, "")
	require.Equal(t, 200, swaggerRes.status)
	require.Contains(t, swaggerRes.body, "API Reference")
}

func TestEnvelopeUsesMessageField(t *testing.T) {
	h := newTestHarness(t)

	loginRes := h.rawRequest("POST", "/api/admin/auth/login", map[string]any{
		"username": "admin",
		"password": "Admin_123",
	}, "")
	require.Equal(t, 200, loginRes.status)
	require.Contains(t, loginRes.body, "\"message\"")
	require.NotContains(t, loginRes.body, "\"msg\"")

	var env apiEnvelope
	require.NoError(t, json.Unmarshal([]byte(loginRes.body), &env))
	require.Equal(t, 0, env.Code)
	require.Equal(t, "OK", env.Message)
}
