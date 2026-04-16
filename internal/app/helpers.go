package app

// helpers.go 仅保留为历史入口，避免继续在单文件中堆叠多职责工具方法。
//
// 原 helpers.go 已按职责拆分到：
// - pagination.go：分页与通用查询工具
// - mask.go：敏感信息脱敏
// - menu_tree.go：菜单树构建
// - auth_session.go：登录会话与权限缓存
// - sms_config.go：短信配置读取与缓存
// - audit.go：审计/登录日志写入与 IP 归属地解析
// - user_lookup.go：用户/用户组查询与软删除用户名处理
// - redis_helpers.go：Redis 字符串/集合操作封装
