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

func Lookup(code string) (BalanceProvider, bool) {
	provider, ok := defaultRegistry[strings.TrimSpace(strings.ToLower(code))]
	return provider, ok
}
