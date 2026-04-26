# 文档入口

本目录记录 MyJob 后端当前真实的架构、模块边界、开发方式、测试方式和迁移背景。

## 按目的阅读

- 快速启动项目：先读 `../README.md`，再读 `development.md`。
- 理解系统定位和能力边界：读 `overview.md`。
- 理解启动链路、分层职责和请求流：读 `architecture.md`。
- 查询业务域、路由、权限和文件归属：读 `module-map.md`。
- 跑测试、对齐 CI 或确认 lint 口径：读 `testing.md`。
- 理解历史迁移背景：读 `migration.md`。
- 查看阶段性规格和实施计划目录约定：读 `superpowers/README.md`；当前已完成文档不再单独保留。
- 查询第三方渠道原始接口资料：读 `../platform_docs/README.md`。

## 维护原则

- 当前事实只写一次：模块归属放在 `module-map.md`，测试策略放在 `testing.md`，开发命令放在 `development.md`。
- Markdown 不重复维护完整字段级 API 文档；字段细节以 `../api/*.go` 和运行时 OpenAPI `/api.json` 为准。
- 代码、路由、测试、CI 或目录结构变化时，同步检查本目录和根 `README.md`。
