package adminlogic

import "myjob/internal/app"

// ProductTemplateLogic 提供商品模板管理相关业务能力。
type ProductTemplateLogic struct{ core *app.Core }

// product_template.go 仅保留 ProductTemplateLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - product_template_query.go：列表查询与基础数据加载
// - product_template_write.go：新增/编辑/删除/批量删除写入逻辑
// - product_template_validate.go：入参归一化与校验
// - product_template_options.go：类型/枚举与展示文案
