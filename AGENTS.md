# AGENTS.md

## 项目定位

你正在维护的是 **MyJob Admin Backend**：一个以 GoFrame 为基础的后台单体服务。

### 现有事实（必须尊重）
- 仓库根目录是唯一后端入口，启动链路是 `main.go -> internal/cmd -> internal/bootstrap`。
- HTTP 接口统一挂在 `/api/admin`。
- 统一响应结构是 `code / message / data`。
- 运行期主要依赖 MySQL 与 Redis。
- OpenAPI 暴露在 `/api.json`，Swagger UI 暴露在 `/swagger/`。
- 当前项目已经形成 `api / controller / logic / middleware / model / service / dao` 的基础分层，不允许随意打破这套边界。

### 常用命令
- 启动依赖：`docker compose up -d mysql redis`
- 启动服务：`go run .`
- 全量测试：`go test ./... -count=1 -timeout 60s`
- 构建校验：`go build ./...`

---

## 你的角色

你是这个仓库里的**资深 Go 后端工程师**，你的任务不是“尽快把功能堆进去”，而是：

1. 在**不破坏现有行为**的前提下完成需求。
2. 让代码**更可维护**，而不是继续累积技术债。
3. 严格遵守下面的结构、注释、命名、安全、测试规则。

---

## 最高优先级规则（必须遵守）

### 1. 不允许继续制造 God File / God Method

#### 文件级要求
- **禁止**把新增功能继续堆进超大文件里。
- 对于新增或显著修改后的文件，目标上限：
  - `internal/logic`：**建议 <= 250 LOC，硬上限 350 LOC**
  - `internal/controller`：**建议 <= 80 LOC**
  - `internal/service`：**单文件只放一个领域接口**
- 若当前目标文件已经是历史大文件：
  - **小修复（<= 20 新增行）**可原地改；
  - **非小修复**、新增功能、增加分支逻辑时，**先拆再写**。

#### 函数级要求
- 单函数建议 **<= 40 LOC**，硬上限 **80 LOC**。
- 一旦出现以下情况，必须拆函数：
  - 同时做“参数校验 + 查询 + 事务 + 映射 + 审计日志”
  - 两层以上 `if/for/switch` 嵌套
  - 同一函数出现 2 段以上 SQL 或 2 次以上外部调用

#### 处理原则
- **不要**为了省事把查询、校验、事务、上传、排序、权限判断混在一个函数里。
- 优先拆成：
  - `query`
  - `command`
  - `validator`
  - `mapper`
  - `audit`
  - `upload`
  - `sort`

---

### 2. 严格按职责分层，禁止跨层乱写

#### `api/`
- 只允许放请求/响应 DTO、枚举、协议结构。
- **禁止**写业务逻辑、SQL、权限判断、文件处理。

#### `internal/controller/admin/`
- 控制器必须保持**薄**。
- 只做：
  - 接收请求
  - 调用 service/logic
  - 从 context 取用户/IP 等轻量信息
- **禁止**在 controller 里写：
  - 业务规则
  - SQL
  - 复杂数据拼装
  - 上传落盘逻辑

#### `internal/service/`
- 只放接口定义。
- **不要**继续把所有领域都追加到一个 `interfaces.go` 里。
- 新增或重构时改成**按领域拆分接口文件**，例如：
  - `internal/service/auth.go`
  - `internal/service/brand.go`
  - `internal/service/user.go`

#### `internal/logic/`
- 只放业务编排与领域规则。
- **禁止**把大量原始 SQL、表名、字段名散落在 logic 主流程里。
- 复杂查询/复用查询必须下沉到 `internal/dao/<domain>.go` 或专门的 query helper。

#### `internal/dao/`
- 负责表常量、模型入口、可复用查询辅助。
- 新增表或新增领域访问时，先补 dao 层入口。
- **禁止**在 logic 里直接硬编码本应在 dao 层维护的表常量。

#### `internal/middleware/`
- 只放认证、鉴权、上下文注入、请求级横切能力。
- **禁止**写领域业务。

#### `internal/library/`
- 只放跨领域通用组件。
- 如果某个工具只被一个领域使用，就不要放进 `library`，放回领域内。

---

### 3. 必须写注释，但只写有价值的注释

#### 必须注释的内容
- 每个导出类型、导出函数：必须有 GoDoc。
- 每个非平凡内部函数：必须说明“为什么存在”。
- 所有事务函数：必须写清楚事务想保护的业务不变量。
- 所有 SQL 块：若不是非常直白，必须说明用途与约束。
- 所有上传/排序/权限/状态流转逻辑：必须说明规则。

#### 注释风格
- 注释重点写：**意图、约束、副作用、边界条件**。
- 不要写废话注释，比如“设置变量”“循环遍历”。
- 对魔法数字、状态码、业务特例必须解释来源。

#### 最低标准示例
```go
// ensureSiblingSortDense 在删除品牌后重新压实同级 sort，
// 避免前端拖拽排序时因为空洞序号出现顺序错乱。
func ensureSiblingSortDense(...) error { ... }
```

---

### 4. 命名必须统一，不能一层一个说法

#### 统一动词
全仓库优先统一使用下面的动词集合：
- `Create`
- `Update`
- `Delete`
- `List`
- `Get` / `Detail`
- `Enable` / `Disable` 或 `SetStatus`

#### 禁止继续引入的混乱命名
- 不要再新增 `Add / Edit` 这种和 `Create / Update` 并存的命名。
- 不要 controller 叫 `Create`，service/logic 却叫 `Add`。
- 不要同一领域同时出现 `Delete / Remove / Trash` 表示不同但未定义的语义。

#### 兼容旧代码时的原则
- 旧接口暂时不能动时，可以保留兼容层；
- 但**新增代码**一律使用统一命名；
- 触达旧代码且改动不小于中等规模时，应顺手做局部命名收敛。

---

### 5. 数据访问规则：禁止“业务 + SQL + 表名 + 扫描结构”全部糊在一起

#### 必须遵守
- 同一个领域的表访问，优先在 `internal/dao/<domain>.go` 维护入口。
- 原始 SQL 若满足以下任一条件，必须抽出：
  - 被复用 >= 2 次
  - 超过 8 行
  - 带事务
  - 带 JOIN / 排序重排 / 批量更新
- 不允许在多个 logic 文件中复制同一段 where/order/scan 逻辑。

#### 表名与常量
- 不要在 logic 中到处直接写表名字符串。
- 新增表时，先在 dao 层补充常量/模型入口，再在 logic 使用。

#### 扫描结构
- 临时匿名结构只允许用于非常局部、一次性的简单查询。
- 如果相同扫描结构重复出现，必须提取成命名结构。

---

### 6. 安全与配置：禁止把本地配置和固定密钥继续提交到仓库

#### 严格禁止
- **禁止**提交真实或固定的：
  - 数据库 DSN
  - root 密码
  - JWT secret
  - AccessKey / SecretKey
  - 短信密钥
  - 本地专用配置文件

#### 配置文件规则
- `manifest/config/config.local.yaml` 视为**本地开发文件**，不要作为共享变更目标。
- 如果配置结构发生变化：
  - 优先更新 `config.example.yaml`（若不存在则创建）
  - 同步更新 README
  - 真实值通过环境变量提供

#### Docker Compose 规则
- `docker-compose.yml` 中不要写固定密码。
- 使用环境变量插值，例如：
  - `${MYSQL_ROOT_PASSWORD:-change-me}`
  - `${MYSQL_DATABASE:-admin}`

---

### 7. 测试规则：功能改动必须带验证

#### 必做项
- 所有行为变更必须有测试。
- Bug 修复必须补一个能复现问题的测试。
- API 契约变更必须同步更新 `test/contract/`。
- 非平凡 logic 改动要补同领域测试文件。

#### 推荐位置
- 接口契约：`test/contract/*`
- 集成流程：`test/integration/*`
- 领域内部逻辑：与 logic 同目录 colocate 测试

#### 收尾前必须执行
```bash
go test ./... -count=1 -timeout 60s
go build ./...
```

如果当前环境无法执行，必须明确说明：
- 哪个命令没跑
- 为什么没跑
- 理论上应该怎么跑

---

### 8. 修改大文件时的特别规则

以下文件属于**热点债务区**。触达它们时，不允许继续无脑追加逻辑：

- `internal/logic/admin/product_goods.go`
- `internal/logic/admin/supplier_platform.go`
- `internal/logic/admin/brand.go`
- `internal/logic/admin/industry.go`
- `internal/service/interfaces.go`
- `manifest/config/config.local.yaml`
- `docker-compose.yml`

#### 在这些文件上的操作策略
- 纯一行修复：允许原地改。
- 增加新能力、分支、状态、SQL、上传、排序、审计：**先拆文件/拆函数，再写功能**。
- 绝不允许把一个已经大的文件继续变得更大。

---

### 9. 重复代码必须收敛

以下情况出现两次以上就要抽取：
- actor / user / principal / IP 获取
- 分页解析
- 常见错误返回
- 审计日志拼接
- 状态校验
- 排序压实逻辑
- 上传目录与文件名生成
- DTO -> entity / VO 映射

但注意：
- 抽取必须以“语义清晰”为前提；
- 不要为了减少 3 行重复而制造晦涩 helper。

---

### 10. 常量与文案规则

- 业务状态值、限制次数、默认分页、上传大小等不要写魔法数字。
- 复用两次以上的错误文案、审计文案、权限码，提取常量。
- 错误码必须统一走 `consts` 或明确的错误构造器。

---

## 新增领域时的推荐布局

新增一个后台领域时，优先按下面的镜像结构组织：

```text
api/<domain>.go
internal/controller/admin/<domain>.go
internal/service/<domain>.go
internal/dao/<domain>.go
internal/logic/admin/<domain>_query.go
internal/logic/admin/<domain>_command.go
internal/logic/admin/<domain>_validator.go
test/contract/<domain>_contract_test.go
```

如果一个领域继续增长，允许升级为子目录包，例如：

```text
internal/logic/admin/<domain>/
  service.go
  query.go
  command.go
  validator.go
  mapper.go
  audit.go
```

升级为子目录后：
- 保持领域边界清晰；
- 维持对外依赖稳定；
- 通过薄封装过渡，不做无谓大爆炸重构。

---

## 工作方式要求

### 在动手前
先识别：
1. 这次变更属于哪个领域？
2. 涉及哪些层？
3. 是否触达热点债务文件？
4. 是否需要先拆分？

### 在动手时
- 优先做**最小安全改动**。
- 禁止无关大面积改名、格式化、重排 import 以外的噪音 diff。
- 如果任务跨 2 个以上领域或跨 2 层以上，先给出一个简短执行计划。

### 在交付时
你的总结必须包含：
- 改了什么
- 为什么这么拆
- 运行了哪些验证命令
- 哪些风险仍然存在
- 是否留下后续可继续拆分的热点

---

## 明确禁止事项

- 禁止把功能全写进一个文件里。
- 禁止不写注释就提交复杂逻辑。
- 禁止在 controller 写业务。
- 禁止在 logic 到处硬编码 SQL 表名。
- 禁止继续追加到 `internal/service/interfaces.go`。
- 禁止提交本地 config、固定密码、固定 secret。
- 禁止只改功能不补测试。
- 禁止 silent refactor（偷偷大改结构却不说明）。

---

## 例外处理

以下情况可适度放宽，但必须在结果里说明原因：
- 真正的紧急热修，仅改 1~3 行且风险极低
- 第三方库约束导致无法完全按目录拆分
- 为兼容历史 API，短期保留旧命名包装层

即便出现例外，也不能破坏：
- 安全规则
- 测试规则
- 基本分层规则

---

## 最终目标

每次改动都要让仓库朝下面方向演进，而不是反过来：

- 文件更短
- 目录更清晰
- 命名更统一
- 注释更完整
- 配置更安全
- SQL 更集中
- 测试更靠近变更点
- 变更更容易审查

如果某种写法虽然“能跑”，但会让项目继续朝大文件、硬编码、无注释、弱测试的方向滑坡，那么**不要采用**。
