package orderlogic

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	runtimeapp "myjob/internal/app"
	"myjob/internal/model/entity"

	"github.com/shopspring/decimal"
)

func (l *OrderLogic) httpClientForOrderProvider(providerCode string) *http.Client {
	if providerCode != ProviderCodeKakayun {
		return l.httpClient
	}
	client := *l.httpClient
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if baseTransport, ok := l.httpClient.Transport.(*http.Transport); ok && baseTransport != nil {
		transport = baseTransport.Clone()
	}
	transport.DisableCompression = true
	client.Transport = transport
	return &client
}

func snapshotOrderRequest(request *http.Request, account entity.SupplierPlatformAccount) (string, error) {
	body := []byte{}
	if request.Body != nil {
		raw, err := io.ReadAll(request.Body)
		if err != nil {
			return "", err
		}
		body = raw
		request.Body = io.NopCloser(bytes.NewReader(raw))
		request.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(raw)), nil
		}
	}
	headers := make(map[string][]string, len(request.Header))
	for key, values := range request.Header {
		copied := make([]string, len(values))
		copy(copied, values)
		headers[key] = copied
	}
	payload := map[string]any{
		"url":     request.URL.String(),
		"method":  request.Method,
		"headers": headers,
		"body":    string(body),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return truncateOrderSnapshot(sanitizeOrderSnapshot(string(raw), account)), nil
}

func sanitizeOrderSnapshot(value string, account entity.SupplierPlatformAccount) string {
	value = strings.ReplaceAll(value, account.SecretKey, runtimeapp.MaskSecret(account.SecretKey))
	value = strings.ReplaceAll(value, account.TokenID, runtimeapp.MaskSecret(account.TokenID))
	return value
}

func truncateOrderSnapshot(value string) string {
	if len(value) <= 4096 {
		return value
	}
	return value[:4096]
}

func defaultOrderMessage(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}

func intToString(value int) string {
	return decimal.NewFromInt(int64(value)).String()
}
