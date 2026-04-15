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

// Lookup 根据 provider_code 查找余额查询适配器实现。
func Lookup(code string) (BalanceProvider, bool) {
	provider, ok := defaultRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}
