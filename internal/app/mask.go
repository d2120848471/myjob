package app

import "myjob/internal/library/region"

// MaskPhone 对手机号进行脱敏展示（保留前后若干位）。
func MaskPhone(phone string) string { return region.MaskPhone(phone) }

// MaskSecret 对密钥/口令类敏感信息进行脱敏展示。
func MaskSecret(value string) string { return region.MaskSecret(value) }

// MaskAccessKey 对 access_key 类敏感信息进行脱敏展示。
func MaskAccessKey(value string) string { return region.MaskAccessKey(value) }
