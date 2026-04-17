package adminlogic

import "myjob/internal/app"

// ProductGoodsChannelBindingLogic 提供商品渠道绑定（绑定级）相关业务能力。
type ProductGoodsChannelBindingLogic struct{ core *app.Core }

// product_goods_channel_binding.go 仅保留 ProductGoodsChannelBindingLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - product_goods_channel_binding_query.go：绑定列表查询与展示字段装配
// - product_goods_channel_binding_write.go：新增/编辑/删除/批量操作/一键排序/自动改价写入逻辑
// - product_goods_channel_binding_validate.go：入参校验、唯一性检查、税态换算与成本价计算
// - product_goods_channel_binding_mapper.go：展示名等字段拼装
