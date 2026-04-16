package adminlogic

import "myjob/internal/app"

// IndustryLogic 提供行业管理及行业-品牌关联管理相关业务能力。
type IndustryLogic struct{ core *app.Core }

// industry.go 仅保留 IndustryLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - industry_query.go：行业与行业品牌关联的查询能力
// - industry_write.go：行业新增/编辑/删除/排序写入逻辑
// - industry_brand_write.go：行业-品牌关联的新增/删除/排序写入逻辑
// - industry_validate.go：行业关联品牌的约束校验
// - industry_mapper.go：排序回写等“映射/重建”辅助逻辑
