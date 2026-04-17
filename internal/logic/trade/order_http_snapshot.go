package tradelogic

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"myjob/internal/app"
)

func snapshotHeaders(headers http.Header, tokenID string, secretKey string) (string, error) {
	copied := make(map[string][]string, len(headers))
	for key, values := range headers {
		items := make([]string, len(values))
		copy(items, values)
		copied[key] = items
	}
	raw, err := json.Marshal(copied)
	if err != nil {
		return "", err
	}
	return truncateSnapshot(sanitizeSnapshot(string(raw), tokenID, secretKey)), nil
}

func snapshotRequestBody(request *http.Request, tokenID string, secretKey string) (string, error) {
	if request == nil {
		return "", nil
	}
	if request.GetBody == nil {
		return "", nil
	}
	reader, err := request.GetBody()
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		return "", err
	}
	return truncateSnapshot(sanitizeSnapshot(string(body), tokenID, secretKey)), nil
}

func snapshotResponseBody(body []byte, tokenID string, secretKey string) string {
	return truncateSnapshot(sanitizeSnapshot(string(body), tokenID, secretKey))
}

func sanitizeSnapshot(value string, tokenID string, secretKey string) string {
	if strings.TrimSpace(secretKey) != "" {
		value = strings.ReplaceAll(value, secretKey, app.MaskSecret(secretKey))
	}
	if strings.TrimSpace(tokenID) != "" {
		value = strings.ReplaceAll(value, tokenID, app.MaskSecret(tokenID))
	}
	return value
}

func truncateSnapshot(value string) string {
	if len(value) <= 4096 {
		return value
	}
	return value[:4096]
}

func shouldRetrySupplierCandidate(response *http.Response, body []byte, parseErr error) bool {
	if parseErr == nil || response == nil {
		return false
	}
	contentType := strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Type")))
	trimmedBody := strings.ToLower(strings.TrimSpace(string(body)))
	if strings.Contains(contentType, "text/html") {
		return true
	}
	return strings.HasPrefix(trimmedBody, "<!doctype html") || strings.HasPrefix(trimmedBody, "<html")
}
