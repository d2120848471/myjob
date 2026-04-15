# AGENTS.md

## 作用
本文件是 **myjob** 仓库的项目级别约束。你的目标不是“最快把功能塞进去”，而是在**不破坏现有架构边界**的前提下，交付**可维护、可测试、可审查、可回滚**的改动。

## 仓库认知
- 这是 Go 后端项目，主入口为 `main.go -> internal/cmd -> internal/bootstrap`。
- 管理端接口根前缀为 `/api/admin`。
- 接口响应遵循统一结构：`code / message / data`。
- `api/` 目录保持**扁平协议目录**；不要重新引入历史式深层嵌套协议包。

## 当前已知高风险文件（默认禁止继续堆业务）
以下文件已经表现出“容易吸附无关逻辑”的风险。**除非只是极小型机械修复，否则不要继续直接把新功能堆进去**：

- `internal/logic/admin/brand.go`
- `internal/logic/admin/industry.go`
- `internal/app/helpers.go`
- `internal/logic/admin/common.go`
- `internal/controller/admin/helper.go`
- `api/common.go`

处理规则：
1. 任务涉及这些文件时，优先在**同 package** 下新建更具体的文件，再做抽离或新增。
2. 新增逻辑时，默认选择按职责拆到 `${domain}_query.go`、`${domain}_write.go`、`${domain}_validate.go`、`${domain}_mapper.go`、`${domain}_options.go` 等文件。
3. **禁止继续扩张** `common.go`、`helper.go`、`helpers.go`、`util.go`、`utils.go`、`misc.go`、`tmp.go` 这类垃圾桶文件名。
4. 若你认为某段代码必须放进上述通用文件，必须先满足两个条件：
   - 该逻辑确实被多个业务域复用，而不是“暂时找不到地方放”；
   - 你在最终说明中明确写出复用对象与理由。

## 优先参考的正向拆分范式
新增或重构时，优先模仿仓库里已经存在的良好拆分方式，而不是再造一套风格：

- `internal/logic/admin/product_goods_{query,write,validate,mapper,options}.go`
- `internal/logic/admin/supplier_platform_{query,write,validate,mapper,balance}.go`

**优先同包多文件拆分，不要为了拆文件而新建无意义子目录。**

## 层级职责（硬约束）

### `api/`
只放：
- 请求/响应结构体
- 路由协议定义
- 与协议相关的常量、枚举、别名

禁止：
- 业务规则
- 数据库访问
- 服务编排
- 与某个具体接口无关的杂项 helper

如果一个类型只服务某一个业务域，就放到对应域文件，不要放进 `api/common.go`。

### `internal/controller/`
只做：
- HTTP/协议参数接收
- 调用 service/logic
- 返回统一响应

禁止：
- 在 controller 内写业务判断、权限分支、组装大段领域逻辑
- 把 controller 写成第二个 logic 层
- 在 controller 内写跨接口复用的通用工具并继续膨胀 `helper.go`

### `internal/service/`
只放：
- 服务接口定义
- 按业务域拆分的接口文件

禁止：
- 把所有接口重新塞回 `interfaces.go`
- 在 service 层写实现

命名要求：
- 一个业务域一个接口文件，优先采用领域名，例如 `brand.go`、`industry.go`、`sms_config.go`
- 不要新建 `service/common.go` 或重新集中到超大接口文件

### `internal/logic/`
只做：
- 业务编排
- 规则校验
- 数据聚合
- 事务边界
- 调用 DAO / model / library

要求：
- 一个业务域内，按职责拆文件，不要把查询、写入、校验、映射、选项组装混写在一个大文件
- 当同一业务文件同时出现“查询 + 写入 + 校验 + DTO 映射 + 选项装配”时，应优先拆分
- 允许同一个 type 在多个文件中实现方法，只要仍在同一 package 下且职责清晰

### `internal/app/` 与 `internal/bootstrap/`
只做：
- 应用装配
- 启动初始化
- 资源注册
- 运行期基础设施拼装

禁止：
- 把具体业务规则、接口专属逻辑、领域判断塞进这里
- 继续扩张 `internal/app/helpers.go` 作为杂项仓库

若改动发生在 app/bootstrap：
- 优先按能力拆分，例如 `http_server.go`、`config_loader.go`、`provider_registry.go`、`upload_bootstrap.go`
- 不要把“顺手写的小工具”继续塞进 `helpers.go`

### `internal/dao/` / `internal/model/`
保持生成物和数据访问职责清晰，禁止把业务编排写入 DAO。

## 文件拆分规则（硬约束）
1. **先按职责判断，再按行数预警。**
2. 即使文件行数不多，只要职责混杂，也必须拆。
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
5. 仅当目录层级本身就是新的稳定边界时，才允许新增子目录；否则优先保持同 package 多文件。

## 注释要求（硬约束）
### 必须写注释的场景
1. **导出类型、导出函数、导出方法**：必须有 Go 风格注释，说明用途，而不是重复名字。
2. **非显而易见的业务规则**：必须解释“为什么”。例如：
   - 为什么需要短信校验
   - 为什么 IP 变化触发校验
   - 为什么某角色可见/不可见
   - 为什么某供应商余额/状态要特殊处理
   - 为什么某字段允许回退或兼容旧值
3. **事务、锁、幂等、缓存、Redis Key、过期时间、回滚策略**：必须解释边界和意图。
4. **外部依赖耦合点**：例如上传目录语义、短信 provider、第三方平台、配置回退链路，必须标注。
5. **临时兼容逻辑**：必须注明兼容对象、移除条件或风险。

### 注释禁止事项
- 禁止把代码逐行翻译成注释。
- 禁止空洞注释，例如“处理数据”“执行逻辑”“查询列表”。
- 禁止用注释掩盖糟糕命名；应优先改名。

## 业务域拆分补充规则
### settings/config 相关
当前“设置”语义下容易混入多种子域。若改动同时涉及短信设置、系统设置、上传设置等，优先拆成更具体文件，例如：
- `settings_sms.go`
- `settings_system.go`
- `settings_upload.go`

不要长期维持一个“settings 大杂烩文件”。

### common 别名/协议
`api/common.go` 只允许保留真正跨多个业务域复用的通用协议别名。任何明显只服务单个域的结构，都必须移回对应域文件。

## 测试与验证（硬约束）
1. 任何非平凡改动，至少补一种：
   - 单元测试
   - 合约测试
   - 集成测试
2. 修 bug 时，优先补能复现该 bug 的测试。
3. 改协议、鉴权、路由、响应结构时，优先考虑合约测试。
4. 改业务编排、分支规则、边界条件时，优先补单测或集成测试。
5. 交付前最少执行：
   - `go test ./... -count=1`
   - `go build ./...`
6. 如果因为外部依赖无法执行完整验证，必须在最终说明中明确：
   - 哪些没跑
   - 为什么没跑
   - 风险落点在哪里

## 文档同步（硬约束）
以下变化发生时，必须同步更新文档：
- 模块边界变化
- 目录结构变化
- 启动方式变化
- 配置项变化
- 测试命令变化
- 新增重要业务流或约束

优先检查并同步：
- `README.md`
- `docs/module-map.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/testing.md`

## 输出要求（强制执行）
### 开始改代码前
先给出一个简短实施说明，至少包含：
- 本次改动所属业务域
- 落在哪一层
- 计划修改/新增哪些文件
- 计划补哪些测试
- 是否需要更新文档

### 完成改动后
必须明确说明：
- 改了哪些文件
- 为什么这些文件归属合理
- 做了哪些拆分
- 加了哪些关键注释
- 跑了哪些验证命令
- 哪些验证没有跑以及原因

## 自检清单
在提交前逐项自检：
- [ ] 是否把不相关职责塞进了同一文件？
- [ ] 是否继续扩大了 `common.go` / `helper.go` / `helpers.go`？
- [ ] controller 是否仍然足够薄？
- [ ] service 是否仍只有接口定义？
- [ ] logic 是否按查询/写入/校验/映射拆清？
- [ ] 导出符号和关键业务规则是否有有效注释？
- [ ] 是否补了足够的测试？
- [ ] 是否同步更新了 README / docs？
- [ ] 如果触碰了已知高风险文件，是否先做了抽离而不是继续堆积？

## 冲突决策
当“快速交付”和“结构清晰”冲突时：
- 优先保证层次边界正确
- 优先保证文件职责单一
- 优先保证可读性和可测试性
- 宁可多一个清晰的小文件，也不要多一个继续膨胀的大文件

任何情况下，都不要为了省几分钟，把新功能直接塞进现有大杂烩文件。
