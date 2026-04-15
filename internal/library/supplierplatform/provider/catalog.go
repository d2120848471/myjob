package supplierprovider

// DefaultTypes 返回系统内置的第三方平台类型字典（用于启动时 seed）。
func DefaultTypes() []PlatformType {
	return []PlatformType{
		{ID: 6, TypeName: "云发卡", ProviderCode: "kakayun", Status: 1, Sort: 6},
		{ID: 7, TypeName: "同系统", ProviderCode: "youkayun", Status: 1, Sort: 7},
		{ID: 15, TypeName: "星海", ProviderCode: "xinghai", Status: 1, Sort: 15},
		{ID: 35, TypeName: "星权益", ProviderCode: "xingquanyi", Status: 1, Sort: 35},
		{ID: 56, TypeName: "雅兰芳", ProviderCode: "feisuyuan", Status: 1, Sort: 56},
		{ID: 72, TypeName: "卡速售", ProviderCode: "kasushou", Status: 1, Sort: 72},
		{ID: 73, TypeName: "卡易信", ProviderCode: "kayixin", Status: 1, Sort: 73},
		{ID: 81, TypeName: "聚浪云", ProviderCode: "julangyun", Status: 1, Sort: 81},
	}
}
