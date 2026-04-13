# 测试说明

## 当前测试分层

### 1. 契约测试

契约测试位于 `test/contract/`，当前是最主要的接口回归入口。
它通过 `NewTestApplication()` 启动应用，底层使用：

- 临时 SQLite 文件
- `miniredis`
- `mock` 短信 sender

当前已覆盖的重点包括：

- 扁平 `api/*.go` 协议目录约束
- 禁止继续引用历史 历史嵌套协议包路径
- OpenAPI / Swagger 暴露
- 统一响应 `message` 字段
- 登录 / 短信二验 / `me` / 退出登录
- 品牌三级结构、行业关联约束与本地上传主流程
- 商品模板验证方式枚举、列表筛选、增删改、批删与非法 ID 校验
- 员工、用户组、主体、短信配置、系统参数配置、日志查询主流程
- 短信发送失败时 Redis 清理行为

运行命令：

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 2. 集成测试

集成测试位于 `test/integration/`。
当前这里只有一个 runtime smoke test，不是完整的外部依赖回归套件。

它当前验证的是：

- 按真实配置创建应用
- 启动 HTTP server
- 使用超级管理员密码完成一次 `/api/admin/auth/login` 请求
- 确认接口能返回成功响应

它需要显式开启：

```bash
export MYJOB_RUN_INTEGRATION=1
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go test ./test/integration -count=1 -timeout 60s
```

如果没有设置环境变量，这个测试会跳过，而不是失败。

### 3. 包内测试

当前仓库里已经有的包内测试主要集中在基础能力层，例如：

- `internal/library/region`
- `internal/library/sms`

适用于纯逻辑或基础库的回归验证。

## 推荐执行顺序

### 日常改动

```bash
go test ./... -count=1 -timeout 60s
```

### 影响接口兼容时

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 影响真实配置或启动链路时

```bash
export MYJOB_RUN_INTEGRATION=1
export SUPER_ADMIN_PHONE=13800000000
export SUPER_ADMIN_PASSWORD=Admin_123
go test ./test/integration -count=1 -timeout 60s
```

## 当前认知边界

目前文档只描述已经存在的测试覆盖，不把“建议补的测试”写成“已经覆盖”。

当前尚未形成完整外部依赖闭环测试集的内容包括但不限于：

- 更细的 MySQL / Redis 行为断言
- 短信 provider 的真实发送链路回归
- 更完整的日志落库和跨重启验证

这些仍属于后续可继续扩充的范围，而不是当前已有事实。
