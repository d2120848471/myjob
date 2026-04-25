package supplierprovider

import "strings"

var defaultRegistry = map[string]BalanceProvider{
	"kakayun":    kakayunProvider{},
	"kayixin":    kayixinProvider{},
	"kasushou":   kasushouProvider{},
	"xingquanyi": xingquanyiProvider{},
	"youkayun":   youkayunProvider{},
	"julangyun":  julangyunProvider{},
	"feisuyuan":  feisuyuanProvider{},
	"xinghai":    xinghaiProvider{},
}

var defaultOrderRegistry = map[string]OrderProvider{
	"kakayun": kakayunProvider{},
}

var defaultProductInfoRegistry = map[string]ProductInfoProvider{
	"kakayun": kakayunProvider{},
}

// Lookup 根据 provider_code 查找余额查询适配器实现。
func Lookup(code string) (BalanceProvider, bool) {
	provider, ok := defaultRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}

// LookupOrder 根据 provider_code 查找下单/查单适配器实现。
func LookupOrder(code string) (OrderProvider, bool) {
	provider, ok := defaultOrderRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}

// LookupProductInfo 根据 provider_code 查找商品详情查询适配器实现。
func LookupProductInfo(code string) (ProductInfoProvider, bool) {
	provider, ok := defaultProductInfoRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}
