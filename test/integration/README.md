# Integration Tests

这里放真实 MySQL / Redis 依赖联动测试。

建议通过环境变量显式启用，例如：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

适合放的内容：

- 启动配置加载
- 数据库与 Redis 连通性
- 超级管理员引导
- 会话、验证码、日志写入闭环
