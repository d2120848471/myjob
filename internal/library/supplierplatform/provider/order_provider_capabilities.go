package supplierprovider

// Capabilities 声明卡卡云下单 provider 的防亏损字段口径。
func (kakayunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "maxmoney"}}
}

// Capabilities 声明卡易信下单 provider 的防亏损字段口径。
func (kayixinProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "safePrice"}}
}

// Capabilities 声明卡速售下单 provider 的防亏损字段口径。
func (kasushouProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "safe_price"}}
}

// Capabilities 声明星权益下单 provider 的防亏损字段口径。
func (xingquanyiProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnit, FieldName: "safe_cost"}}
}

// Capabilities 声明优卡云下单 provider 的防亏损字段口径。
func (youkayunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "maxmoney"}}
}

// Capabilities 声明聚浪云下单 provider 的防亏损字段口径。
func (julangyunProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeTotal, FieldName: "accessPrice"}}
}

// Capabilities 声明星海下单 provider 的防亏损字段口径。
func (xinghaiProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnit, FieldName: "itemPrice"}}
}

// Capabilities 声明飞速源只能单数量提交，且不支持上游防亏损字段。
func (feisuyuanProvider) Capabilities() OrderProviderCapabilities {
	return OrderProviderCapabilities{MaxQuantityPerCreate: 1, SafetyPrice: SafetyPriceCapability{Mode: SafetyPriceModeUnsupported}}
}
