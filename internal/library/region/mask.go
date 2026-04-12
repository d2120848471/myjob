package region

import "strings"

func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

func MaskAccessKey(value string) string {
	return maskMiddle(value, 4, 4)
}

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
