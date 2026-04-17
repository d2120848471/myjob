package supplierprovider

import "strings"

var defaultOrderRegistry = map[string]OrderProvider{
	"youkayun": youkayunProvider{},
}

// LookupOrder 根据 provider_code 查找下单/查单适配器实现。
func LookupOrder(code string) (OrderProvider, bool) {
	provider, ok := defaultOrderRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}
