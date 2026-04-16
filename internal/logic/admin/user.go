package adminlogic

import "myjob/internal/app"

// UserLogic 提供员工账号管理相关业务能力。
type UserLogic struct{ core *app.Core }

// user.go 仅保留 UserLogic 声明，避免继续在单文件中堆叠多职责逻辑。
//
// 具体实现已按职责拆分到：
// - user_query.go：列表与回收站查询
// - user_write.go：新增/编辑/删除/恢复/启停写入逻辑
// - user_notify.go：余额通知开关
// - user_business.go：批量设置/取消商务逻辑
