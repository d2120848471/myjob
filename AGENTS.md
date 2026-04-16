# AGENTS.md

## 作用
本文件是 **myjob** 仓库的项目级别约束。目标不是“最快把功能塞进去”，而是在 **不破坏当前边界** 的前提下，交付 **可维护、可测试、可审查、可回滚** 的改动。

---

## 仓库认知
- 这是 Go 后端项目，主入口为 `main.go -> internal/cmd -> internal/bootstrap`。
- 管理端接口根前缀为 `/api/admin`。
- 接口响应遵循统一结构：`code / message / data`。
- `api/` 保持 **扁平协议目录**；不要重新引入历史式深层嵌套协议包。
- 默认优先选择 **同 package 多文件拆分**，而不是为了“按功能分文件夹”去新建没有稳定边界的子目录。
- 只有当目录本身代表 **长期稳定的包边界/能力边界** 时，才允许新增子目录，例如 provider、specs、plans 这类目录。

---

## 当前仓库现状（先理解，再动手）

### 已经拆好的正向范式（继续沿用，不要回退）
以下模式说明仓库已经从“大文件堆逻辑”转向“同包多文件按职责拆分”：

- `internal/logic/admin/brand.go` 仅保留 `BrandLogic` 声明；具体实现分散在：
  - `brand_query.go`
  - `brand_write.go`
  - `brand_upload.go`
  - `brand_validate.go`
  - `brand_mapper.go`
- `internal/logic/admin/industry.go` 仅保留 `IndustryLogic` 声明；具体实现分散在：
  - `industry_query.go`
  - `industry_write.go`
  - `industry_brand_write.go`
  - `industry_validate.go`
  - `industry_mapper.go`
- `internal/logic/admin/product_template.go` 已拆到 `product_template_{query,write,validate,options}.go`
- `internal/logic/admin/purchase_limit.go` 已拆到 `purchase_limit_{query,write,validate,options}.go`
- `internal/logic/admin/user.go` 已拆到 `user_{query,write,notify,business}.go`
- `internal/app/helpers.go` 仅保留历史入口说明；公共能力已经按职责拆到：
  - `pagination.go`
  - `mask.go`
  - `menu_tree.go`
  - `auth_session.go`
  - `sms_config.go`
  - `audit.go`
  - `user_lookup.go`
  - `redis_helpers.go`
- 继续优先参考：
  - `internal/logic/admin/product_goods_{query,write,validate,mapper,options}.go`
  - `internal/logic/admin/supplier_platform_{query,write,validate,mapper,balance}.go`

**硬约束：** 不要把上述已拆好的域重新揉回一个 `brand.go` / `industry.go` / `helpers.go` / `settings.go` 大文件。

### 当前真实高风险点（默认禁止继续堆业务）
以下文件或文件名仍然容易继续吸附无关职责。除非只是极小型机械修复，否则不要继续直接往里堆：

- `api/common.go`
- `api/settings.go`
- `internal/controller/admin/settings.go`
- `internal/controller/admin/helper.go`
- `internal/logic/admin/common.go`（只允许保留 service 聚合和极小型公共胶水）
- 任何新的 `common.go` / `helper.go` / `helpers.go` / `util.go` / `utils.go` / `misc.go` / `tmp.go`

处理规则：
1. 任务涉及这些文件时，优先在 **同 package** 下新增更具体的文件，再做抽离或迁移。
2. 如果某个通用文件只是在“暂时找不到地方放”，那它就不应该继续增长。
3. 若你坚持把代码放进通用文件，必须同时满足：
   - 该逻辑被多个域稳定复用；
   - 你在最终说明中明确写出复用对象与理由。

### 与契约测试绑定的文件名（特别注意）
`test/contract/api_layout_test.go` 当前会显式检查 `api/` 目录中的若干文件名（例如 `settings.go`、`common.go`）。

因此：
- 若只是把 `settings.go`、`common.go` 内部职责拆开，**优先保留原文件作为薄入口/说明文件**；
- 若确实要删除、改名或改变协议文件集合，必须 **同步更新**：
  - `test/contract/api_layout_test.go`
  - `test/contract/README.md`
  - `README.md`
  - `docs/module-map.md`

---

## 层级职责（硬约束）

### `api/`
只放：
- 请求/响应结构体
- 路由协议定义
- 与协议直接相关的常量、枚举、别名

禁止：
- 业务规则
- 数据库访问
- 服务编排
- 与某个接口无关的通用 helper

额外要求：
- 协议仍保持 `api/*.go` 扁平结构，不新建无意义子目录。
- 一个类型如果只服务某个业务域，就放到该域文件，不要塞进 `api/common.go`。
- `api/common.go` 只允许保留 **真正跨多个业务域复用** 的别名，例如通用分页、登录用户等；明显只服务单域的别名必须移回领域文件。

### `internal/controller/`
只做：
- HTTP/协议参数接收
- 调用 service/logic
- 返回统一响应

禁止：
- 在 controller 内写业务判断、权限分支、事务逻辑
- 把 controller 写成第二个 logic 层
- 借 `helper.go` 持续沉淀跨接口杂项逻辑

额外要求：
- controller 按业务域拆文件。
- 若“设置”类接口已经分化为短信设置 / 系统设置 / 上传设置等子域，controller 也应同步拆开，而不是长期维持一个 `settings.go` 大杂烩。

### `internal/service/`
只放：
- 服务接口定义
- 按业务域拆分的接口文件

禁止：
- 把所有接口重新塞回一个超大 `interfaces.go`
- 在 service 层写实现
- 新建 `service/common.go`

### `internal/logic/`
只做：
- 业务编排
- 规则校验
- 数据聚合
- 事务边界
- 调用 DAO / model / library

要求：
- 一个业务域内，按职责拆文件，不要把查询、写入、校验、映射、选项装配混写。
- 允许同一个 type 在多个文件中实现方法，只要仍在同一 package 且职责清晰。
- `internal/logic/admin/common.go` 不承载领域逻辑；默认只允许放公共错误构造、service 聚合、极小型跨域胶水。

### `internal/app/` 与 `internal/bootstrap/`
只做：
- 应用装配
- 启动初始化
- 资源注册
- 运行期基础设施拼装

禁止：
- 把具体业务规则、接口专属逻辑、领域判断塞进这里
- 重新把公共能力堆回 `internal/app/helpers.go`

如果改动发生在 app/bootstrap：
- 优先按能力拆，例如：`config_loader.go`、`provider_registry.go`、`request_meta.go`、`upload_bootstrap.go`
- 不要为了省事再造一个万能 `helpers.go`

### `internal/dao/` / `internal/model/`
- 保持生成物和数据访问职责清晰。
- 禁止把业务编排、权限判断、流程控制写进 DAO。

---

## 文件拆分规则（硬约束）
1. **先按职责判断，再按行数预警。**
2. 即使文件行数不多，只要职责混杂，也应该拆。
3. 文件行数经验线：
   - 目标：`<= 250` 行
   - 警戒：`> 350` 行时，新增前先评估拆分
   - 默认应拆：`> 450` 行时，优先抽离后再继续加功能
   - 红线：`> 600` 行时，除生成文件外，禁止继续直接堆代码
4. 同一业务域新增能力时，优先使用以下命名：
   - `${domain}_query.go`
   - `${domain}_write.go`
   - `${domain}_validate.go`
   - `${domain}_mapper.go`
   - `${domain}_options.go`
   - `${domain}_convert.go`
   - `${domain}_permission.go`
   - `${domain}_upload.go`
   - `${domain}_notify.go`
5. 仅当目录本身就是新的稳定边界时，才允许新增子目录；否则优先保持 **同 package 多文件**。
6. 对外已经被测试/文档/协作约定锚定的文件名（如 `api/settings.go`、`api/common.go`），若只是职责拆分，优先保留原文件为薄入口，而不是粗暴删掉。

---

## 业务域拆分补充规则

### settings / config 相关
当前“设置”语义下容易混入多种子域。

若改动同时涉及短信设置、系统设置、上传设置等，优先拆成更具体文件，例如：
- `settings_sms.go`
- `settings_system.go`
- `settings_upload.go`

适用层：
- `api/`
- `internal/controller/admin/`
- `internal/logic/admin/`（如确有必要）

默认策略：
- `api/settings.go` 可保留为薄入口/说明文件；
- 具体协议定义放到 `api/settings_sms.go`、`api/settings_system.go`；
- `internal/controller/admin/settings.go` 可仅保留 `SettingsController` 声明与构造，具体 handler 分拆到更具体文件。

### common 别名 / 协议
- `api/common.go` 只保留 **真正跨域复用** 的通用别名。
- 任何明显只服务单域的结构，都必须移回对应域文件。
- 不要为了“看起来统一”而把 Brand / Industry / PurchaseLimit / SupplierPlatform / Log 相关别名长期留在 `api/common.go`。

### request meta / client ip 辅助
- `internal/controller/admin/helper.go` 目前只允许保持很小的请求元信息辅助。
- 一旦职责扩展，应改成更清晰的文件名，如 `request_meta.go`、`client_ip.go`，且禁止变成新的杂项垃圾桶。

---

## 注释要求（硬约束）

### 必须写注释的场景
1. **导出类型、导出函数、导出方法**：必须有 Go 风格注释，说明用途，而不是重复名字。
2. **非显而易见的业务规则**：必须解释“为什么”。例如：
   - 为什么需要短信校验
   - 为什么 IP 变化触发校验
   - 为什么某角色可见/不可见
   - 为什么某供应商余额/状态要特殊处理
   - 为什么某字段允许兼容旧写法
3. **事务、锁、幂等、缓存、Redis Key、TTL、回滚策略**：必须解释边界和意图。
4. **外部依赖耦合点**：例如上传目录语义、短信 provider、第三方平台、配置回退链路，必须标注。
5. **临时兼容逻辑**：必须注明兼容对象、移除条件或风险。

### 分层注释要求
#### `api/*.go`
- 所有导出 `Req` / `Res` / `Item` / `Enum` 类型都应有 Go 风格注释。
- 注释应说明该协议服务的场景，不要只是翻译 `summary` 或字段标签。
- 对兼容旧参数（如单组/多组写法兼容）要明确写出兼容含义。

#### `internal/controller/admin/*.go`
- 导出 controller 类型、构造函数、导出 handler 方法都应有注释。
- 注释聚焦“这个 handler 对外暴露什么能力”，不要抄接口路径。

#### `internal/logic/admin/*.go`
- 导出方法必须有注释。
- 对关键校验、排序重排、缓存读写、回滚边界、三方调用降级策略，补充说明“为什么这样做”。

#### `internal/app/*.go`
- 对缓存 key、TTL、provider fallback、会话模型、敏感信息脱敏规则，补充“为什么”的注释。

### 注释禁止事项
- 禁止把代码逐行翻译成注释。
- 禁止空洞注释，例如“处理数据”“执行逻辑”“查询列表”。
- 禁止用注释掩盖糟糕命名；应优先改名或拆分。

---

## 测试与验证（硬约束）
1. 任何非平凡改动，至少补一种：
   - 单元测试
   - 合约测试
   - 集成测试
2. 改协议、鉴权、路由、响应结构、文件布局约束时，优先考虑合约测试。
3. 改业务编排、分支规则、边界条件时，优先补单测或集成测试。
4. 若变更 `api/` 文件布局或文件名，必须检查并更新：
   - `test/contract/api_layout_test.go`
   - `test/contract/README.md`
5. 交付前最少执行：
   - `go test ./... -count=1 -timeout 60s`
   - `go build ./...`
   - `golangci-lint run --timeout=5m`
6. 如果因为外部依赖无法执行完整验证，必须在最终说明中明确：
   - 哪些没跑
   - 为什么没跑
   - 风险落点在哪里

---

## 文档同步（硬约束）
以下变化发生时，必须同步更新文档：
- 模块边界变化
- 目录结构变化
- 启动方式变化
- 配置项变化
- 测试命令变化
- CI / lint 约束变化
- 新增重要业务流或约束

优先检查并同步：
- `README.md`
- `docs/overview.md`
- `docs/module-map.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/testing.md`
- `docs/migration.md`
- `docs/superpowers/**`（若本次改动对应已有 spec / plan 或新增 spec / plan）
- `test/contract/README.md`
- `test/integration/README.md`

额外要求：
- 若 README 中存在目录树或文档索引，必须与当前仓库真实结构一致。
- 若仓库 CI 已包含 lint，但 README / docs 只写了 test/build，则需要补齐 lint 说明。
- 若 README / docs 仍把已拆分域写成单文件（例如 `brand.go`、`industry.go`、`user.go`、`product_template.go`、`purchase_limit.go`），应改成能反映现状的写法，例如 `brand*.go`、`user*.go`，或直接列出新的拆分文件。

---

## 本轮默认优先级（Codex / Agent 执行时优先处理）
### P0
- 把 `api/settings.go` 内的短信设置 / 系统设置协议拆开（优先保留 `settings.go` 为薄入口，同时新增 `settings_sms.go`、`settings_system.go`）。
- 把 `internal/controller/admin/settings.go` 的短信设置 / 系统设置 handler 拆开。
- 收缩 `api/common.go`，仅保留真正跨域复用的别名，其余移回各自领域文件。

### P1
- 对 `api/*.go` 做导出协议类型注释扫除。
- 对 `internal/controller/admin/*.go` 做导出 controller / constructor / handler 注释扫除。
- 更新 `README.md`、`docs/module-map.md`、`docs/development.md`、`docs/testing.md`，补齐当前拆分结构与 lint/CI 约束。
- 把 `README.md` 的文档索引补到 `docs/superpowers/**`。

### P2
- 处理 `internal/controller/admin/helper.go` 的命名与注释问题，确保它保持极小且不继续膨胀。
- 审查 `internal/logic/admin/common.go` 是否仍然只承担 service 聚合和通用胶水职责。

---

## 输出要求（强制执行）
### 开始改代码前
先给出一个简短实施说明，至少包含：
- 本次改动所属业务域
- 落在哪一层
- 计划修改 / 新增哪些文件
- 计划补哪些测试
- 计划同步哪些文档

### 完成改动后
必须明确说明：
- 改了哪些文件
- 为什么这些文件归属合理
- 做了哪些拆分
- 哪些旧文件被保留为薄入口，为什么要保留
- 加了哪些关键注释
- 更新了哪些测试 / 文档 / AGENTS 约束
- 跑了哪些验证命令
- 哪些验证没跑以及原因

---

## 自检清单
在提交前逐项自检：
- [ ] 是否把不相关职责塞进了同一文件？
- [ ] 是否继续扩大了 `common.go` / `helper.go` / `helpers.go`？
- [ ] 是否错误地为了“按功能分文件夹”创建了无意义子目录？
- [ ] 是否保持了 `api/` 扁平协议目录？
- [ ] service 是否仍只有接口定义？
- [ ] controller 是否仍然足够薄？
- [ ] logic 是否按查询 / 写入 / 校验 / 映射拆清？
- [ ] `settings` 是否已经按子域拆开？
- [ ] `api/common.go` 是否只保留了真正通用的内容？
- [ ] 导出符号和关键业务规则是否有有效注释？
- [ ] 若触碰了 `api/settings.go` / `api/common.go`，是否同步检查了契约测试？
- [ ] 是否补了足够的测试？
- [ ] 是否同步更新了 README / docs / test README / AGENTS？
- [ ] 是否补齐了 lint 验证说明？

---

## 冲突决策
当“快速交付”和“结构清晰”冲突时：
- 优先保证层次边界正确
- 优先保证文件职责单一
- 优先保证同 package 多文件拆分
- 优先保证可读性和可测试性
- 宁可多一个清晰的小文件，也不要多一个继续膨胀的大文件
- 宁可保留一个被测试锚定的薄入口文件，也不要为了“看起来整洁”粗暴删除并破坏协作约定

任何情况下，都不要为了省几分钟，把新功能直接塞进现有大杂烩文件。
