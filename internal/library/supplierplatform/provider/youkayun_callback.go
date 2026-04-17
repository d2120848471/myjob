package supplierprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func (youkayunProvider) VerifyCallbackSignature(account AccountConfig, headers http.Header, body []byte) error {
	// 优卡云回调文档未提供验签规则，一期按“无签名”处理。
	return nil
}

func (youkayunProvider) ParseCallbackPayload(account AccountConfig, headers http.Header, body []byte) (*CallbackResult, error) {
	fields, err := parseCallbackFields(headers, body)
	if err != nil {
		return nil, err
	}
	providerRequestOrderNo := strings.TrimSpace(fields["orderno"])
	channelOrderNo := strings.TrimSpace(fields["outorderno"])
	status := strings.TrimSpace(fields["status"])

	idempotencyKey := providerRequestOrderNo
	if idempotencyKey == "" {
		idempotencyKey = channelOrderNo
	}
	if idempotencyKey == "" {
		idempotencyKey = md5Lower(string(body))
	}

	return &CallbackResult{
		ProviderRequestOrderNo: providerRequestOrderNo,
		ChannelOrderNo:         channelOrderNo,
		FinalSuccess:           status == "3",
		FinalFailed:            status == "5",
		UpstreamStatus:         status,
		IdempotencyKey:         idempotencyKey,
		RawPayload:             string(body),
	}, nil
}

func (youkayunProvider) BuildCallbackAck(input CallbackAckInput) ([]byte, string, error) {
	payload := map[string]any{
		"code": 1000,
		"msg":  "ok",
	}
	raw, _ := json.Marshal(payload)
	return raw, "application/json", nil
}

func parseCallbackFields(headers http.Header, body []byte) (map[string]string, error) {
	mediaType, params, _ := mime.ParseMediaType(headers.Get("Content-Type"))
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType == "" {
		trimmed := strings.TrimSpace(string(body))
		if strings.HasPrefix(trimmed, "{") {
			return parseJSONFields(body)
		}
		return parseQueryFields(body)
	}

	switch {
	case strings.Contains(mediaType, "application/json"):
		return parseJSONFields(body)
	case strings.Contains(mediaType, "application/x-www-form-urlencoded"):
		return parseQueryFields(body)
	case strings.Contains(mediaType, "multipart/form-data"):
		boundary := strings.TrimSpace(params["boundary"])
		if boundary == "" {
			return map[string]string{}, fmt.Errorf("missing multipart boundary")
		}
		return parseMultipartFields(body, boundary)
	default:
		trimmed := strings.TrimSpace(string(body))
		if strings.HasPrefix(trimmed, "{") {
			return parseJSONFields(body)
		}
		return parseQueryFields(body)
	}
}

func parseJSONFields(body []byte) (map[string]string, error) {
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	fields := make(map[string]string, len(payload))
	for key, value := range payload {
		v := strings.TrimSpace(fmt.Sprint(value))
		if v == "<nil>" {
			v = ""
		}
		fields[strings.TrimSpace(key)] = v
	}
	return fields, nil
}

func parseQueryFields(body []byte) (map[string]string, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	fields := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) == 0 {
			fields[strings.TrimSpace(key)] = ""
			continue
		}
		fields[strings.TrimSpace(key)] = strings.TrimSpace(items[0])
	}
	return fields, nil
}

func parseMultipartFields(body []byte, boundary string) (map[string]string, error) {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	fields := make(map[string]string)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(part)
		_ = part.Close()
		if err != nil {
			return nil, err
		}
		fields[strings.TrimSpace(part.FormName())] = strings.TrimSpace(string(data))
	}
	return fields, nil
}
