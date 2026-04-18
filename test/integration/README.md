# Integration Tests

这里放需要真实配置参与的联动测试。

## 当前现状

当前目录包括：

- runtime smoke test（`runtime_smoke_test.go`）：需要真实配置参与，验证应用能启动并完成一次 `/api/admin/auth/login` 请求
- supplier 平台余额刷新集成回归（`supplier_platform_balance_test.go`）：验证主/备域名请求策略、HTTP 降级、余额日志落库等行为
  - 文件内还包含一个可选的 live provider 验证用例，用环境变量显式开启

它目前不是完整的 MySQL / Redis / 短信 / 日志闭环回归集。

## 运行前提

默认运行：

```bash
go test ./test/integration -count=1 -timeout 60s
```

运行 runtime smoke test：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

runtime smoke test 会使用本地默认超管 `admin / abc123` 登录。

运行 supplier live provider 余额验证（可选）：

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_DOMAIN=example.com
export SUPPLIER_LIVE_BACKUP_DOMAIN=example.com
export SUPPLIER_LIVE_TOKEN_ID=1008612345
export SUPPLIER_LIVE_SECRET_KEY=secret
go test ./test/integration -count=1 -timeout 60s
```
