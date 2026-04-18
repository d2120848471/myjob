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
- 第三方对接 OpenAPI 路径暴露、菜单种子同步、平台账号 CRUD、余额刷新与软删重建回归
- 员工、用户组、主体、短信配置、系统参数配置、日志查询主流程
- 短信发送失败时 Redis 清理行为

运行命令：

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 2. 集成测试

集成测试位于 `test/integration/`。
当前已经包含两类入口，不再只有 runtime smoke test。

它当前验证的是：

- 按真实配置创建应用
- 启动 HTTP server
- 使用超级管理员密码完成一次 `/api/admin/auth/login` 请求
- 确认接口能返回成功响应
- 第三方平台余额刷新在主域名 / 备用域名 / `https` 降级下的真实执行顺序
- 业务失败和传输失败的分流、余额日志脱敏与 trace_id 落库

其中 `supplier-platform` 集成测试默认使用 `httptest.Server` 模拟平台，不依赖真实外部账号：

```bash
go test ./test/integration -run 'TestSupplierPlatformRefresh_' -count=1 -timeout 60s
```

它需要显式开启：

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

它会使用本地默认超管 `admin / abc123` 完成登录烟测。

### 2.1 第三方平台 live 验证

第三方平台余额刷新还支持显式开启 live 验证，用真实 `domain/token_id/secret_key` 跑一次 `/api/admin/supplier-platforms/{id}/balance/refresh`：

```bash
export MYJOB_RUN_SUPPLIER_LIVE=1
export SUPPLIER_LIVE_TYPE_ID=35
export SUPPLIER_LIVE_NAME='木木（星权益未税）'
export SUPPLIER_LIVE_DOMAIN=xqy.api.xqy1.cn
export SUPPLIER_LIVE_BACKUP_DOMAIN=xqy.api.xqy1.cn
export SUPPLIER_LIVE_TOKEN_ID=74
export SUPPLIER_LIVE_SECRET_KEY='***'
go test ./test/integration -run TestSupplierPlatformRefresh_LiveProviderBalance -count=1 -v
```

这个入口只在显式设置环境变量时执行，适合做单平台连通性核验，不作为默认 CI 套件。

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

并建议在提交前执行一次 lint（与 CI 对齐）：

```bash
golangci-lint run --timeout=5m
```

CI 里会使用增量参数（`--new-from-rev=origin/main`）减少无关历史问题的干扰。

### 影响接口兼容时

```bash
go test ./test/contract -count=1 -timeout 60s
```

### 影响真实配置或启动链路时

```bash
export MYJOB_RUN_INTEGRATION=1
go test ./test/integration -count=1 -timeout 60s
```

## 当前认知边界

目前文档只描述已经存在的测试覆盖，不把“建议补的测试”写成“已经覆盖”。

当前尚未形成完整外部依赖闭环测试集的内容包括但不限于：

- 更细的 MySQL / Redis 行为断言
- 短信 provider 的真实发送链路回归
- 更完整的跨重启验证和多平台批量 live 回归

这些仍属于后续可继续扩充的范围，而不是当前已有事实。
