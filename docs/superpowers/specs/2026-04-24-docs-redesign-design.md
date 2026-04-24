# 文档体系重构设计

日期：2026-04-24

## 背景

当前仓库的主文档已经覆盖了启动、架构、模块、测试和迁移背景，但存在两类问题：

- 文档入口和真实目录不完全一致，例如 `README.md` 与 `docs/module-map.md` 引用了不存在的 `docs/superpowers/README.md`。
- 多篇文档重复维护当前功能、目录清单和测试覆盖，后续新增功能时容易出现局部更新、整体漂移。

本次目标是重写文档体系，而不是修改业务代码。文档应以当前真实仓库为准，保持简洁、可查、可维护。

## 外部参考

本设计参考以下成熟文档体系，并只吸收适合当前 Go 后端单体仓库的部分：

- Diátaxis：按教程、How-to、参考、解释拆分文档职责。
- Kubernetes 文档：用文档入口组织 Concepts、Tasks、Reference 等阅读路径，并要求代码变化同步文档。
- Django 文档：把教程、使用指南、How-to、Reference 分开，减少单篇文档承担过多职责。
- FastAPI 文档：入门路径可执行，API/Reference 单独维护。
- Go 官方文档：源码注释和代码强绑定，Markdown 不重复维护过细 API 字段细节。

## 目标

1. 建立清晰的文档入口，让读者能按目的找到对应文档。
2. 收敛重复内容，避免 README、overview、module-map、testing 多处重复维护同一事实。
3. 让文档和当前目录、路由、测试、CI、lint、配置保持一致。
4. 保留必要历史背景，但不让历史文档承担当前事实清单的职责。
5. 为后续 spec 和 plan 提供稳定入口，补齐 `docs/superpowers/README.md`。

## 非目标

- 不改 Go 代码、SQL、配置、CI workflow 或测试逻辑。
- 不引入静态文档站、文档生成器或额外依赖。
- 不在 Markdown 中重复维护完整字段级 API 文档；字段细节仍以 `api/*.go` 和 OpenAPI `/api.json` 为准。
- 不删除已跟踪文档文件。若实施中发现确实需要删除文件，必须单独按危险操作确认机制处理。

## 文档信息架构

### 根 README

`README.md` 是项目首页，只承担以下职责：

- 项目定位和当前运行形态。
- 快速启动路径。
- 最小验证命令。
- 关键边界提示，例如根目录为唯一后端入口、`api/` 扁平协议目录、统一响应结构。
- 文档索引，指向 `docs/README.md` 和主要专题文档。

`README.md` 不再展开完整模块清单、业务域细节和测试覆盖细节。

### 文档门户

新增 `docs/README.md` 作为文档门户，按阅读目的组织入口：

- 想快速跑起来：读 `README.md` 与 `docs/development.md`。
- 想理解系统：读 `docs/overview.md` 与 `docs/architecture.md`。
- 想查模块归属：读 `docs/module-map.md`。
- 想跑测试或对齐 CI：读 `docs/testing.md`。
- 想理解迁移背景：读 `docs/migration.md`。
- 想查看规格或实施计划：读 `docs/superpowers/README.md`。
- 想查第三方渠道原始接口：读 `platform_docs/README.md`。

### 当前能力概览

`docs/overview.md` 作为解释型文档，描述项目是什么、当前已实现什么、不做什么。

它应保留业务域概览和非目标，但不维护每个业务域的文件列表、测试命令或完整路由表。

### 架构说明

`docs/architecture.md` 作为解释型文档，描述：

- 启动链路。
- 分层职责。
- 请求处理链路。
- 认证、权限、短信、订单 worker 等关键运行流程。
- MySQL、Redis、短信 provider、审计、第三方 provider 等运行时依赖。

它不维护完整业务模块清单，避免和 `docs/module-map.md` 重复。

### 模块地图

`docs/module-map.md` 作为参考型文档，是业务域对照的唯一主文档。每个业务域统一描述：

- 协议文件。
- controller 文件。
- service 接口。
- logic 文件或文件模式。
- 路由前缀。
- 权限边界。
- 核心能力和明确限制。

它可以维护目录地图和路由权限摘要，但不写本地启动步骤、CI 说明或迁移故事。

### 开发说明

`docs/development.md` 作为 How-to 文档，保留可执行步骤：

- 本地依赖。
- 启动 MySQL / Redis。
- 配置来源。
- 启动应用。
- 常用命令。
- SQL/schema 同步规则。
- DAO 生成规则。

它只写开发者要执行的动作和必要解释，不重复业务域能力清单。

### 测试说明

`docs/testing.md` 作为 How-to + Reference 文档，统一维护：

- 测试分层。
- 默认测试命令。
- 契约测试、集成测试、包内测试的适用场景。
- CI test/build/lint 口径。
- 外部依赖和 live 验证开关。
- 当前测试认知边界。

`test/contract/README.md` 与 `test/integration/README.md` 只保留目录内最小运行说明，详细策略回链到 `docs/testing.md`。

### 迁移说明

`docs/migration.md` 作为历史背景文档，压缩并保留：

- 为什么从历史结构迁到当前根应用。
- 已完成的关键迁移方向。
- 当前仍需遵守的兼容面。

它不再维护完整当前协议文件清单和目录清单，这些事实归 `docs/module-map.md` 管。

### Superpowers 文档入口

新增 `docs/superpowers/README.md`，说明该目录用于存放需求设计、实施计划和阶段性规格文档。

同时新增或保留子目录约定：

- `docs/superpowers/specs/`：设计规格。
- `docs/superpowers/plans/`：实施计划。

当前设计 spec 放在 `docs/superpowers/specs/2026-04-24-docs-redesign-design.md`。

### 第三方平台原始文档

`platform_docs/README.md` 保持平台原始接口资料总览。它只描述渠道侧资料，不承担本仓库实现说明；本仓库实现归 `docs/module-map.md`、`docs/architecture.md` 和 provider 代码。

## 清理规则

1. 优先重写和收敛重复内容，不删除已跟踪文档。
2. 删除文件、批量移动文件或改变文档目录结构前，必须单独确认。
3. Markdown 中的目录树必须和 `git ls-files` 能看到的真实结构一致。
4. 对 OpenAPI 已能表达的字段细节，只在 Markdown 中写边界和入口，不逐字段重复。
5. 历史背景和当前事实分离：当前事实只维护在 README、docs portal、module-map、development、testing 等当前文档中。

## 实施影响面

计划修改或新增：

- `README.md`
- `docs/README.md`
- `docs/overview.md`
- `docs/architecture.md`
- `docs/module-map.md`
- `docs/development.md`
- `docs/testing.md`
- `docs/migration.md`
- `docs/superpowers/README.md`
- `test/contract/README.md`
- `test/integration/README.md`

计划保持不改或仅在必要时微调：

- `platform_docs/README.md`
- `platform_docs/*.md`
- Go 源码、SQL、配置、CI workflow、测试代码

## 验证方案

文档重构完成后至少执行：

- `go test ./test/contract -run TestAPIProtocolLayout -count=1 -timeout 60s`：确认被文档强调的 API 文件布局约束仍然真实。
- `go test ./test/contract -run TestCIWorkflow -count=1 -timeout 60s`：确认 CI/lint 文档口径和 workflow 约束一致。
- `go test ./... -count=1 -timeout 60s`：全量回归，确保文档变更没有破坏嵌入式路径或契约检查。
- `go build ./...`：构建验证。
- `golangci-lint run --timeout=5m`：与本地文档推荐命令一致。

如果本地环境缺少 MySQL、Redis 或 `golangci-lint`，最终说明必须明确未运行项、原因和风险落点。

## 验收标准

- `README.md` 不再引用不存在的路径。
- `docs/README.md` 能作为稳定文档门户使用。
- `docs/superpowers/README.md` 存在，并解释 specs/plans 目录用途。
- `docs/migration.md` 不再维护容易过时的完整当前文件清单。
- `docs/module-map.md` 是业务域、路由和权限归属的唯一主参考。
- `docs/testing.md` 是测试策略和 CI/lint 口径的唯一主参考。
- `test/contract/README.md`、`test/integration/README.md` 保持轻量，不重复整套测试策略。
- 文档不描述尚未实现的功能，不把“建议补充”写成“已经覆盖”。

## 风险与应对

- 风险：重写文档时引入新的事实错误。应对：实施时先用 `rg`、`git ls-files`、OpenAPI/路由代码和测试 README 交叉核对。
- 风险：README 过度收缩导致入口信息不足。应对：保留快速启动、验证命令和清晰文档索引。
- 风险：module-map 继续膨胀。应对：只写参考事实和边界，不写开发流程、测试策略和历史背景。
- 风险：测试命令依赖本地 MySQL 或 lint 工具。应对：最终交付时明确实际运行结果与未运行原因。
