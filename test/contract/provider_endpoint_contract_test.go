package contract_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func xingquanyiSignContract(secret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	builder.WriteString(secret)
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString(params[key])
	}
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func TestProviderCallbackEndpoint_ReturnsRawAck(t *testing.T) {
	h := newTestHarness(t)
	ctx := context.Background()
	now := h.app.Core().Now()

	subjectResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	_, err = h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('渠道账号A', 'youkayun', '优卡云', 7, ?, 0, 'http://example.com', 'token-a', 'secret', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)

	body := map[string]any{
		"userid":     "token-a",
		"orderno":    "PR-001",
		"outorderno": "CH-001",
		"status":     3,
	}
	res := h.rawRequest("POST", "/api/provider/youkayun/order-callback", body, "")
	require.Equal(t, 200, res.status)
	require.Contains(t, res.body, `"code":1000`)
}

func TestProviderPriceNotifyEndpoint_ReturnsOkBody(t *testing.T) {
	h := newTestHarness(t)
	ctx := context.Background()
	now := h.app.Core().Now()

	subjectResult, err := h.app.Core().DB().Exec(ctx, `
INSERT INTO admin_subject (name, has_tax, created_at, updated_at)
VALUES ('交易主体A', 1, ?, ?)
`, now, now)
	require.NoError(t, err)
	subjectID, _ := subjectResult.LastInsertId()

	_, err = h.app.Core().DB().Exec(ctx, `
INSERT INTO supplier_platform_account (name, provider_code, provider_name, type_id, subject_id, has_tax, domain, token_id, secret_key, created_at, updated_at)
VALUES ('星权益账号', 'xingquanyi', '星权益', 35, ?, 1, 'xqy.example.com', '6', 'secretXYZ', ?, ?)
`, subjectID, now, now)
	require.NoError(t, err)

	params := map[string]string{
		"customer_id":  "6",
		"timestamp":    strconv.FormatInt(now.Unix(), 10),
		"event_type":   "price_changed",
		"event_data":   `{"price":"11.000"}`,
		"product_id":   "2",
		"product_name": "爱奇艺黄金会员年卡",
		"product_type": "1",
	}
	payload := map[string]any{
		"customer_id":  6,
		"timestamp":    now.Unix(),
		"event_type":   params["event_type"],
		"event_data":   params["event_data"],
		"product_id":   2,
		"product_name": params["product_name"],
		"product_type": 1,
	}
	payload["sign"] = xingquanyiSignContract("secretXYZ", params)

	res := h.rawRequest("POST", "/api/provider/xingquanyi/price-notify", payload, "")
	require.Equal(t, 200, res.status)
	require.Contains(t, strings.ToLower(res.body), "ok")

	// 幂等重复请求仍应直接 ACK。
	time.Sleep(10 * time.Millisecond)
	res2 := h.rawRequest("POST", "/api/provider/xingquanyi/price-notify", payload, "")
	require.Equal(t, 200, res2.status)
	require.Contains(t, strings.ToLower(res2.body), "ok")
}
