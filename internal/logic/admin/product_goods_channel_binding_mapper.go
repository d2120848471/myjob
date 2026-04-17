package adminlogic

import "strings"

func buildBindingDisplayName(supplierGoodsName, supplierGoodsNo, subjectName, providerName string) string {
	base := strings.TrimSpace(supplierGoodsName)
	if base == "" {
		base = strings.TrimSpace(supplierGoodsNo)
	}
	subjectName = strings.TrimSpace(subjectName)
	providerName = strings.TrimSpace(providerName)
	return strings.TrimSpace(base + " / " + subjectName + " / " + providerName)
}
