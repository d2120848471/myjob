package supplierprovider

import "strings"

var defaultPriceNotifyRegistry = map[string]PriceNotifyProvider{
	"xingquanyi": xingquanyiProvider{},
}

// LookupPriceNotify 根据 provider_code 查找价格通知适配器实现。
func LookupPriceNotify(code string) (PriceNotifyProvider, bool) {
	provider, ok := defaultPriceNotifyRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}

