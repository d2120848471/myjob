package api

import "github.com/gogf/gf/v2/frame/g"

// SettingsSystemGetReq 用于读取系统参数配置。
//
// Group 为空时返回全部分组；Group 非空时仅返回指定分组。
type SettingsSystemGetReq struct {
	g.Meta `path:"/settings/system" method:"get" tags:"设置" summary:"读取系统参数配置" security:"BearerAuth" dc:"按分组读取系统参数配置"`
	Group  string `json:"group" dc:"配置分组，为空时返回全部分组"`
}

// SettingsSystemItem 表示一个系统参数项（组内键 + 展示信息 + 当前值）。
type SettingsSystemItem struct {
	Key        string `json:"key" dc:"组内键名"`
	Label      string `json:"label" dc:"展示名称"`
	Value      string `json:"value" dc:"参数值"`
	ValueType  string `json:"value_type" dc:"参数类型"`
	Unit       string `json:"unit,omitempty" dc:"单位"`
	Required   bool   `json:"required" dc:"是否必填"`
	Configured bool   `json:"configured" dc:"是否已配置"`
	UpdatedAt  string `json:"updated_at,omitempty" dc:"更新时间"`
}

// SettingsSystemGroup 表示一个系统参数分组及其包含的参数项列表。
type SettingsSystemGroup struct {
	Group string               `json:"group" dc:"配置分组"`
	Label string               `json:"label,omitempty" dc:"分组名称"`
	Items []SettingsSystemItem `json:"items" dc:"参数列表"`
}

// SettingsSystemGetRes 返回系统参数配置。
//
// - 单组读取时会填充 Group/Label/Items，并同时返回 Groups（仅包含该组），便于前端统一处理
// - 全量读取时仅返回 Groups
type SettingsSystemGetRes struct {
	Group  string                `json:"group,omitempty" dc:"配置分组"`
	Label  string                `json:"label,omitempty" dc:"分组名称"`
	Items  []SettingsSystemItem  `json:"items,omitempty" dc:"参数列表"`
	Groups []SettingsSystemGroup `json:"groups,omitempty" dc:"分组参数列表"`
}

// SettingsSystemSaveReq 用于保存系统参数配置，支持单组与多组两种写法。
//
// 兼容旧单组写法：
// - 当 Groups 为空时，使用 Group + Items 作为单组保存输入
// - 当 Groups 非空时，以 Groups 为准（优先多组批量保存）
type SettingsSystemSaveReq struct {
	g.Meta `path:"/settings/system" method:"put" tags:"设置" summary:"保存系统参数配置" security:"BearerAuth" dc:"按分组批量保存系统参数配置"`
	Group  string                    `json:"group" dc:"配置分组，兼容旧单组写法"`
	Items  []SettingsSystemSaveItem  `json:"items" dc:"参数列表，兼容旧单组写法"`
	Groups []SettingsSystemSaveGroup `json:"groups" dc:"分组参数列表"`
}

// SettingsSystemSaveItem 表示一个需要保存的系统参数键值对。
type SettingsSystemSaveItem struct {
	Key   string `json:"key" v:"required#key不能为空" dc:"组内键名"`
	Value string `json:"value" dc:"参数值"`
}

// SettingsSystemSaveGroup 表示一个需要保存的系统参数分组输入。
type SettingsSystemSaveGroup struct {
	Group string                   `json:"group" v:"required#group不能为空" dc:"配置分组"`
	Items []SettingsSystemSaveItem `json:"items" v:"required#items不能为空" dc:"参数列表"`
}

// SettingsSystemSaveRes 表示系统参数保存成功（返回体为空）。
type SettingsSystemSaveRes struct{}
