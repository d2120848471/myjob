# 测试说明

## 推荐默认命令

```bash
go test ./... -count=1 -timeout 60s
go build ./...
golangci-lint run --timeout=5m
```

`go test ./...` 会覆盖包内测试、契约测试和默认可运行的集成测试。部分 live 或真实配置测试需要显式环境变量。

## 契约测试

目录：`test/contract/`

用途：

- 约束扁平 `api/*.go` 协议目录。
- 防止回退到历史嵌套协议包。
- 验证 OpenAPI / Swagger 暴露。
- 验证统一响应 `code / message / data`。
- 覆盖登录、短信二验、权限、主要后台业务、开放订单和后台订单主流程。

运行：

```bash
go test ./test/contract -count=1 -timeout 60s
```

契约测试通过 `NewTestApplication()` 启动测试态应用，底层使用 MySQL 测试库 `admin_test`、`miniredis` 和 mock 短信 sender。

## 集成测试

目录：`test/integration/`

默认可运行的集成测试使用 `httptest.Server` 模拟第三方平台或云发卡，不依赖真实外部账号。

运行：

```bash
go test ./test/integration -count=1 -timeout 60s
```

订单 worker 聚焦回归：

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

runtime smoke test 需要显式开启：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

第三方平台 live 验证需要真实账号并显式开启：

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_NAME='示例平台'
export SUPPLIER_LIVE_DOMAIN=example.com
export SUPPLIER_LIVE_BACKUP_DOMAIN=example.com
export SUPPLIER_LIVE_TOKEN_ID=1008612345
export SUPPLIER_LIVE_SECRET_KEY=secret
go test ./test/integration -run TestSupplierPlatformRefresh_LiveProviderBalance -count=1 -v
```

## 包内测试

包内测试集中在基础能力和业务逻辑边界，例如：

- `internal/app`
- `internal/library/region`
- `internal/library/sms`
- `internal/library/supplierplatform/provider`
- `internal/logic/admin`
- `internal/logic/order`

schema 和注释约束：

```bash
go test ./internal/app -run 'Test(MySQLSchemaIncludesTableAndColumnComments|ManifestMySQLSchemaFilesIncludeTableAndColumnComments)' -count=1 -timeout 60s
```

云发卡 provider 聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run TestKakayunOrderProvider -count=1 -timeout 60s
```

卡卡云商品信息同步聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductInfoProvider|TestLookupProductInfo' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestSyncChannelBindingsOnce|TestSaveInventoryConfigTriggersImmediateSyncWhenSwitchEnabled|TestProductGoodsChannelSyncWorker' -count=1 -timeout 60s
go test ./internal/bootstrap -run TestApplicationStartsAndClosesBackgroundWorkers -count=1 -timeout 60s
```

卡卡云商品推送订阅和改价记录聚焦回归：

```bash
go test ./internal/library/supplierplatform/provider -run 'TestKakayunProductSubscriptionProvider|TestKakayunProductChangePushProvider|TestLookupProductPush' -count=1 -timeout 60s
go test ./internal/logic/admin -run 'TestAutoSubscribeKakayunBinding|TestSupplierProductSubscription|TestCancel|TestResubscribe|TestApplyProductGoodsChannelPriceChange' -count=1 -timeout 60s
go test ./test/contract -run 'TestOpenSupplierProductChangeCallbackReturnsPlainOK|TestSupplierProductSubscriptionListCancelAndResubscribe|TestProductGoodsChannelPriceChangeList' -count=1 -timeout 60s
```

订单渠道定价和主体快照聚焦回归：

```bash
go test ./internal/library/channelpricing ./internal/app ./test/integration -run 'Test(EffectiveSellPrice|OrderSnapshot|ExternalOrderSchemaContainsRequiredTablesAndIndexes|OrderWorkerUsesSelectedChannelSubjectAndAutoPriceAmounts)' -count=1 -timeout 60s
```

## CI 与 lint

CI workflow 位于 `.github/workflows/ci.yml`：

- `test` job 启动 MySQL 8.4 service，执行 `go test ./... -count=1 -timeout 60s` 和 `go build ./...`。
- `lint` job 执行 `golangci-lint`，参数包含 `--timeout=5m --new-from-rev=origin/main`。

`.golangci.yml` 当前启用：

- `govet`
- `staticcheck`
- `ineffassign`
- `unused`
- `typecheck`

本地建议在提交前执行：

```bash
golangci-lint run --timeout=5m
```

## 常见测试失败排查

### MySQL 连接失败

现象通常是 `dial tcp 127.0.0.1:3306`、认证失败或 `admin_test` 初始化失败。

处理顺序：

1. 确认依赖已启动：`docker compose up -d mysql redis`
2. 确认端口没有被其它 MySQL 占用：`docker compose ps`
3. 重新跑单包测试缩小范围：`go test ./internal/app -count=1 -timeout 60s`
4. 如果库结构异常，重建本地测试依赖后再跑：`docker compose down -v && docker compose up -d mysql redis`

### 测试库锁等待或相互踩数据

`NewTestCore()` 会使用 MySQL 命名锁串行化 `admin_test` 的建库和清库流程。若测试被强制中断，可能出现锁等待到超时。

处理顺序：

1. 先确认没有残留的 `go test` 进程。
2. 重新执行失败包：`go test ./test/integration -count=1 -timeout 60s`
3. 若仍超时，重启 MySQL 容器释放残留连接。

### 契约测试提示 API 布局不符合预期

常见原因是新增了 `api/` 子目录、改名或删除了被契约测试锚定的协议文件。

处理顺序：

1. 查看 `test/contract/api_layout_test.go` 的失败文件名。
2. 若只是职责拆分，优先保留原文件为薄入口。
3. 若确实改变协议文件集合，同步更新 `test/contract/README.md`、`README.md`、`docs/module-map.md`。
4. 复跑：`go test ./test/contract -run TestAPIProtocolLayout -count=1 -timeout 60s`

### schema 或 SQL 注释测试失败

常见原因是只改了 `internal/app/schema.go` 或只改了 `manifest/sql/*.sql`，两边没有同步。

处理顺序：

1. 对照失败提示补齐表注释、字段注释或索引定义。
2. 同步修改 `manifest/sql/*.sql` 和 `internal/app/schema.go`。
3. 复跑：

```bash
go test ./internal/app -run 'Test(MySQLSchemaIncludesTableAndColumnComments|ManifestMySQLSchemaFilesIncludeTableAndColumnComments)' -count=1 -timeout 60s
```

### 订单 worker 测试失败

先看失败点是提交、轮询、补单还是恢复异常提交状态。订单 worker 测试依赖 `httptest.Server` 模拟上游，不应访问真实第三方平台。

处理顺序：

1. 聚焦复现：`go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s -v`
2. 单独运行失败用例，判断是否和前置用例顺序相关。
3. 若单独通过、组合失败，优先排查全局状态、测试库清理、时间精度和后台 goroutine。
4. 修改后必须复跑完整 `test/integration`，不能只看单个用例。

### live 测试被跳过或失败

live 测试默认跳过是预期行为。只有设置 `MYJOB_RUN_SUPPLIER_LIVE=1` 且提供真实账号环境变量时，才会访问外部平台。

处理顺序：

1. 确认是否真的需要 live 验证。
2. 检查 `SUPPLIER_LIVE_*` 环境变量是否完整。
3. 失败时先确认账号、域名和上游平台状态，不要把 live 失败直接等同于本仓库逻辑错误。

### lint 命令不存在

本地出现 `golangci-lint: command not found` 表示当前机器没有安装 lint 工具。

处理顺序：

1. 记录该项未运行，不要写成 lint 通过。
2. 仍需执行 `go test ./... -count=1 -timeout 60s` 和 `go build ./...`。
3. 依赖 CI 的 lint job 做最终校验，或在本机安装与 CI 兼容的 `golangci-lint` 后复跑。

## 当前认知边界

文档只描述已经存在的测试覆盖，不把建议补充项写成已经覆盖。

当前尚未形成完整外部依赖闭环测试集的内容包括：

- 真实短信发送链路回归。
- 跨重启行为验证。
- 多平台批量 live 回归。
- 更细的 MySQL / Redis 行为断言。
