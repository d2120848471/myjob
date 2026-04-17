package adminlogic

import "myjob/internal/app"

// ProductGoodsChannelConfigLogic 提供商品渠道配置（商品级）相关业务能力。
type ProductGoodsChannelConfigLogic struct{ core *app.Core }

// product_goods_channel_config.go 仅保留 ProductGoodsChannelConfigLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - product_goods_channel_config_query.go：读取与默认配置初始化
// - product_goods_channel_config_write.go：更新与摘要刷新
// - product_goods_channel_config_validate.go：入参校验与归一化
