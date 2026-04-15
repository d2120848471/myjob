package region

import "strings"

// MaskPhone 对 11 位手机号进行脱敏（中间 4 位替换为 *）。
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskAccessKey 对 access_key 类字段进行脱敏展示。
func MaskAccessKey(value string) string {
	return maskMiddle(value, 4, 4)
}

// MaskSecret 对 secret/密钥类字段进行脱敏展示。
func MaskSecret(value string) string {
	return maskMiddle(value, 4, 4)
}

func maskMiddle(value string, prefix, suffix int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= prefix+suffix {
		return "****"
	}
	maskSize := len(value) - prefix - suffix
	if maskSize < 4 {
		maskSize = 4
	}
	return value[:prefix] + strings.Repeat("*", maskSize) + value[len(value)-suffix:]
}
