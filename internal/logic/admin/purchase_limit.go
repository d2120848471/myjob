package adminlogic

import "myjob/internal/app"

// PurchaseLimitLogic 提供商品购买数量限制策略管理相关业务能力。
type PurchaseLimitLogic struct{ core *app.Core }

// purchase_limit.go 仅保留 PurchaseLimitLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - purchase_limit_query.go：列表查询与基础数据加载
// - purchase_limit_write.go：新增/编辑/删除/状态切换写入逻辑
// - purchase_limit_validate.go：入参归一化与校验
// - purchase_limit_options.go：枚举/字典与展示文案
