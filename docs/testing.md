# 测试说明

## 测试分层

### 1. 就近单元测试

适用于纯逻辑或基础能力：

- `internal/logic/admin/*_test.go`
- `internal/library/*/*_test.go`
- 需要测试包内细节时，可配合 `testdata/`

### 2. 契约测试

放在 `test/contract/`，用于验证对外接口兼容性和核心业务流：

- 登录
- 短信发送 / 验证
- `me`
- 退出登录
- 用户、用户组、主体、短信配置、日志查询等关键流程

当前契约测试入口是：

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 3. 集成测试

放在 `test/integration/`，只验证真实依赖联动：

- MySQL / Redis 连通性
- 根应用配置加载
- 超级管理员引导
- 会话、短信验证码、日志落库等闭环

这类测试默认通过环境开关显式开启，避免开发机每次全量测试都强依赖外部服务。

## 默认验收命令

```bash
go test ./... -count=1 -timeout 60s
go build ./...
```

## Fixture 约定

- `test/fixture/`：跨包共享的测试样例和说明
- `testdata/`：只服务于单个包的本地样例数据

## 建议执行顺序

1. 改动基础逻辑时先跑就近单元测试
2. 影响接口兼容时跑 `test/contract`
3. 修改配置、SQL、缓存、日志链路时再跑 `test/integration`
