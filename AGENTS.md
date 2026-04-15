# AGENTS.md

## 目标
本仓库是 GoFrame 单体后台项目。你在这个仓库里工作的首要目标不是“尽快把功能堆出来”，而是在**不破坏现有分层和契约**的前提下，交付可维护、可测试、可追踪的改动。

默认要求：
- 优先小步修改，避免无关重构。
- 优先保持现有架构边界清晰。
- 优先补测试和文档，而不是只改代码让它“看起来能跑”。

---

## 仓库架构认知（必须遵守）

### 入口与运行时
- 入口在仓库根：`main.go -> internal/cmd -> internal/bootstrap`
- 这是当前唯一需要维护和发布的后台入口。

### 目录职责
- `api/`：**只放请求/响应协议定义**，不放业务逻辑、不放数据库访问、不放服务编排。
- `internal/controller/admin/`：**只做 HTTP 协议适配**，包括收参、取登录态、取客户端 IP、调用 service、返回结果；不要把业务规则写进 controller。
- `internal/service/`：**只定义接口边界**，controller 只能依赖这里暴露的抽象。
- `internal/logic/admin/`：**业务编排层**，负责规则、校验、事务、状态流转、跨组件协调。
- `internal/app/`：运行时核心能力与核心依赖封装。
- `internal/library/`：跨模块通用基础能力。
- `internal/bootstrap/`：应用组装、路由注册、中间件挂载；不要把业务细节写到这里。
- `manifest/`、`docs/`、`test/`：配置、文档、测试资源。

### 已有契约
- HTTP 接口统一挂在 `/api/admin`
- 统一响应格式是 `code / message / data`
- `api/` 当前采用**扁平 `api/*.go` 协议目录**；**不要重新引入历史嵌套协议包路径**。

---

## 强约束：禁止行为

### 1. 禁止把不相关职责塞进一个文件
出现下面任一情况时，必须拆文件，而不是继续往原文件追加：
- 同一文件同时包含“查询 + 写入 + 校验 + 映射 + 常量 + DTO 适配 + 第三方调用”多种职责
- 同一业务域文件已经明显过长，还继续往里面追加新功能
- 为了省事，把多个独立接口、多个业务分支、多个辅助类型全部堆进同一个 `.go` 文件

### 2. 禁止把业务逻辑写进非业务层
- `api/` 里禁止出现业务规则、数据库访问、service 调用
- `controller` 里禁止写 SQL、事务、复杂权限规则、复杂数据整形
- `bootstrap` 里禁止写具体业务流程

### 3. 禁止无语义命名
禁止新增或扩展以下“垃圾桶”文件，除非内容真的具有跨模块通用性：
- `common.go`
- `helper.go`
- `util.go`
- `misc.go`
- `tmp.go`

如果只是某个领域专用逻辑，必须用**领域前缀**命名文件，而不是继续往 `common.go` / `helper.go` 里堆。

### 4. 禁止只改代码不补说明
以下改动必须同步补充说明：
- 新增/修改公开接口
- 新增/修改关键业务规则
- 新增/修改环境变量、配置项、脚本、启动方式
- 新增/修改测试入口或测试前置条件

### 5. 禁止“为通过而通过”的测试策略
- 不要为了让测试通过而删除断言、弱化断言或跳过关键路径
- 不要把该补的测试偷换成口头说明
- 不要把明显应该在单测覆盖的业务规则，全部推给集成测试或人工验证

---

## 文件拆分规则（非常重要）

### 总原则
本项目允许在**同一 package 下按业务域拆成多个文件**。不要为了“分模块”盲目新建很多目录；优先遵守当前包结构，在**同包内按职责拆文件**。

### 推荐拆分方式
以领域前缀拆分，例如：
- `product_goods_query.go`
- `product_goods_write.go`
- `product_goods_validate.go`
- `product_goods_mapper.go`
- `product_goods_options.go`
- `supplier_platform_query.go`
- `supplier_platform_write.go`
- `supplier_platform_balance.go`
- `supplier_platform_provider.go`

### `internal/service/` 特别规则
- 不要继续把所有 service 接口都堆在 `interfaces.go`
- 按领域拆成多个文件，但保留 `package service`，例如：
  - `auth.go`
  - `user.go`
  - `group.go`
  - `product_goods.go`
  - `supplier_platform.go`

### 文件长度红线
以下是建议红线，不是语法限制，但你必须主动遵守：
- 目标：单文件尽量控制在 **300 行以内**
- 警戒：超过 **500 行** 时，默认应拆分
- 红线：超过 **700 行** 的业务文件，除非是生成文件或纯协议/纯枚举，否则应视为设计问题

如果你正在修改一个已超长文件，优先先做**顺手拆分**，再继续加功能。

---

## 注释与文档规则

### 注释不是装饰，而是约束
不要写“翻译代码”的废话注释；要写能解释**为什么这样做**的注释。

### 必须补注释的场景
- 导出的类型、函数、方法：如果其用途对包外调用者不够直观，必须有 Go 风格注释
- 复杂业务规则：首登二验、IP 变化二验、权限边界、super-only 规则、状态机、库存/余额/限购等规则
- Redis key、缓存语义、锁语义、验证码/会话清理策略
- 事务边界、幂等边界、失败回滚原因
- 第三方平台或供应商适配中的特殊约束
- 任何魔法值、兼容性分支、历史包袱处理

### 不要这样写注释
- 不要逐行复述代码
- 不要给显而易见的 getter / setter / 简单赋值写注释
- 不要为了“看起来很规范”而堆大量空洞注释

### `api/*.go` 特别说明
`api/*.go` 里的 `g.Meta`、`summary`、`dc` 本身就是接口文档的一部分。这里优先维护这些元信息；只有在字段或语义明显不直观时，再补额外注释。

---

## 业务层实现约束

### Controller 层
Controller 只允许做这些事：
- 接收请求
- 调用 service
- 取当前用户 / principal / client IP
- 返回结果

Controller 不应做这些事：
- 拼 SQL
- 开事务
- 直接写表
- 大段 if/else 业务规则
- 第三方平台调用细节

### Logic 层
Logic 负责：
- 参数归一化
- 业务校验
- 权限和状态规则
- 事务编排
- 调用 DAO / app / library / provider
- 输出接口层需要的数据

### Bootstrap 层
Bootstrap 只做装配：
- 组装 core
- 组装 service / controller
- 注册路由
- 注册 middleware
- 配 OpenAPI / Swagger

不要把业务判断和业务开关塞到 bootstrap。

---

## 测试策略

### 变更必须带验证
做任何非纯重命名类变更，都要补至少一种验证：
- 业务函数/校验逻辑：优先补就近 `_test.go`
- 接口行为、响应结构、协议约束：补 `test/contract`
- 真实运行态、依赖联动：按需补 `test/integration`

### 运行顺序
默认按下面顺序思考和执行：
1. 先跑与改动最相关的最小测试集
2. 再跑 `go test ./... -count=1 -timeout 60s`
3. 再跑 `go build ./...`

### 集成测试说明
- `test/integration` 是显式环境开关控制的，不要默认假设它总能跑
- 如果你没有运行集成测试，要明确说明**为什么没跑**，而不是假装已经验证

---

## 文档同步规则

只要出现下面任一情况，必须同步更新文档：
- 目录职责变化
- 新增业务模块或明显拆分模块
- 新增/删除接口
- 启动方式变化
- 配置项、环境变量、脚本变化
- 测试命令或测试前提变化

重点关注这些文件：
- `README.md`
- `docs/module-map.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/testing.md`
- `test/contract/README.md`
- `test/integration/README.md`

不要让 README 和真实目录、真实功能长期漂移。

---

## 开发与验证命令

### 常用命令
```bash
docker compose up -d mysql redis
export ADMIN_CONFIG=manifest/config/config.local.yaml
go run .
```

### 基础验证
```bash
go test ./... -count=1 -timeout 60s
go build ./...
```

### 契约测试
```bash
go test ./test/contract -count=1 -timeout 60s
```

### 集成测试（按需）
```bash
export MYJOB_RUN_INTEGRATION=1
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go test ./test/integration -count=1 -timeout 60s
```

---

## 工作方式要求

每次开始改动前，先在心里完成这 5 个判断：
1. 这次改动属于哪个业务域？
2. 应该落在哪一层？`api / controller / service / logic / app / library / bootstrap`
3. 现有文件是否已经过大？是否应该拆到同包的多个文件？
4. 这次改动会不会影响契约测试、README、模块地图？
5. 需要补哪类测试？

如果答案不清楚，不要直接开写；先选择更保守、更符合现有架构边界的方案。

---

## 提交前自检清单

在结束前逐项自检：
- [ ] 没有把业务逻辑写进 `api/`、`controller`、`bootstrap`
- [ ] 没有把多个不相关职责堆进同一文件
- [ ] 新增代码按领域前缀拆文件，而不是继续扩张 `common.go` / `helper.go`
- [ ] 复杂规则有解释性注释
- [ ] `api/*.go` 的 `g.Meta` / `summary` / `dc` 保持准确
- [ ] 相关测试已新增或更新
- [ ] 已运行相关测试 / 构建，或明确说明未运行原因
- [ ] 需要同步的文档已更新
- [ ] 没有顺手做与任务无关的大重构

---

## 完成标准

只有同时满足下面条件，任务才算完成：
- 功能实现正确
- 架构边界没有被破坏
- 文件组织没有继续恶化
- 注释和文档足以解释关键规则
- 测试和构建有可验证结果
- 变更范围可审查、可回滚、可维护

如果“快速堆代码”和“保持结构清晰”冲突，优先选择后者。
