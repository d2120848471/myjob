package supplierprovider

import "net/http"

// CallbackAckInput 是回调 ACK 构建输入。
type CallbackAckInput struct {
	ProviderRequestOrderNo string
	ChannelOrderNo         string
	FinalSuccess           bool
	FinalFailed            bool
}

// CallbackResult 是 Provider 解析回调后的统一结果对象。
type CallbackResult struct {
	ProviderRequestOrderNo string
	ChannelOrderNo         string

	FinalSuccess bool
	FinalFailed  bool

	UpstreamStatus string
	ErrorCategory  string
	ErrorCode      string
	ErrorMessage   string

	IdempotencyKey string
	RawPayload     string
}

// CallbackProvider 定义上游订单回调验签、解析与 ACK 构建能力。
type CallbackProvider interface {
	Code() string
	Name() string
	VerifyCallbackSignature(account AccountConfig, headers http.Header, body []byte) error
	ParseCallbackPayload(account AccountConfig, headers http.Header, body []byte) (*CallbackResult, error)
	BuildCallbackAck(input CallbackAckInput) ([]byte, string, error)
}

