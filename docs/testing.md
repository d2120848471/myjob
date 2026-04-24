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

## 当前认知边界

文档只描述已经存在的测试覆盖，不把建议补充项写成已经覆盖。

当前尚未形成完整外部依赖闭环测试集的内容包括：

- 真实短信发送链路回归。
- 跨重启行为验证。
- 多平台批量 live 回归。
- 更细的 MySQL / Redis 行为断言。
