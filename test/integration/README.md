# Integration Tests

集成测试用于验证需要多组件联动或模拟外部服务的行为。完整测试策略见 `../../docs/testing.md`。

## 默认运行

```bash
go test ./test/integration -count=1 -timeout 60s
```

默认测试使用 `httptest.Server` 模拟第三方平台或云发卡，不依赖真实外部账号。

## 聚焦运行

订单 worker：

```bash
go test ./test/integration -run TestOrderWorker -count=1 -timeout 60s
```

runtime smoke test：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

第三方平台 live 验证：

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
