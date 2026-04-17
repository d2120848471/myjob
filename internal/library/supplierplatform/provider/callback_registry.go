package supplierprovider

import "strings"

var defaultCallbackRegistry = map[string]CallbackProvider{
	"youkayun": youkayunProvider{},
}

// LookupCallback 根据 provider_code 查找回调验签/解析适配器实现。
func LookupCallback(code string) (CallbackProvider, bool) {
	provider, ok := defaultCallbackRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}

