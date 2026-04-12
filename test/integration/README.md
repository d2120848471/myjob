# Integration Tests

这里放需要真实配置参与的联动测试。

## 当前现状

当前目录里只有一个 runtime smoke test。
它验证的是：

- 应用能按真实配置创建并启动
- 能完成一次 `/api/admin/auth/login` 请求
- 接口返回成功响应

它目前不是完整的 MySQL / Redis / 短信 / 日志闭环回归集。

## 运行前提

```bash
export MYJOB_RUN_INTEGRATION=1
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go test ./test/integration -count=1 -timeout 60s
```

如果没有设置这些环境变量，测试会跳过。
