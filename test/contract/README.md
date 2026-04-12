# Contract Tests

这里放接口契约和核心业务流测试。

## 当前覆盖范围

- 扁平 `api/*.go` 协议目录约束
- 禁止继续引用历史 历史嵌套协议包路径
- OpenAPI `/api.json` 和 Swagger `/swagger/` 暴露
- 统一响应字段 `code / message / data`
- 账号密码登录、短信二验、`me`、退出登录
- 员工、用户组、主体、短信配置、日志查询主流程
- 短信发送失败时的 Redis 清理行为

## 当前运行方式

契约测试会启动测试态应用，底层默认使用：

- SQLite 临时文件
- `miniredis`
- mock 短信 sender

运行命令：

```bash
go test ./test/contract -count=1 -timeout 60s
```
