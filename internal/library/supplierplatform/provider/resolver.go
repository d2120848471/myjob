package supplierprovider

import (
	"fmt"
	"strings"
)

// ResolveProvider 根据平台类型或域名/名称信息推断 provider_code 与展示名称。
func ResolveProvider(typeID int, domain, backupDomain, name string) (string, string, error) {
	switch typeID {
	case 6:
		return "kakayun", "卡卡云", nil
	case 7:
		return "youkayun", "优卡云", nil
	case 15:
		return "xinghai", "星海", nil
	case 35:
		return "xingquanyi", "星权益", nil
	case 56:
		return "feisuyuan", "飞速源", nil
	case 72:
		return "kasushou", "卡速售", nil
	case 73:
		return "kayixin", "卡易信", nil
	case 81:
		return "julangyun", "聚浪云", nil
	default:
		normalized := strings.ToLower(strings.Join([]string{domain, backupDomain, name}, " "))
		switch {
		case strings.Contains(normalized, "kakayun"):
			return "kakayun", "卡卡云", nil
		case strings.Contains(normalized, "youkayun"):
			return "youkayun", "优卡云", nil
		case strings.Contains(normalized, "xinghai"):
			return "xinghai", "星海", nil
		case strings.Contains(normalized, "xingquanyi"):
			return "xingquanyi", "星权益", nil
		case strings.Contains(normalized, "feisuyuan"):
			return "feisuyuan", "飞速源", nil
		case strings.Contains(normalized, "kasushou"):
			return "kasushou", "卡速售", nil
		case strings.Contains(normalized, "kayixin"):
			return "kayixin", "卡易信", nil
		case strings.Contains(normalized, "julangyun"):
			return "julangyun", "聚浪云", nil
		default:
			return "", "", fmt.Errorf("平台类型暂不支持")
		}
	}
}
