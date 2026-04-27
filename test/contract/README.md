# Contract Tests

契约测试用于约束接口兼容、协议布局和核心业务流。完整测试策略见 `../../docs/testing.md`。

## 运行

```bash
go test ./test/contract -count=1 -timeout 60s
```

## 当前重点

- 扁平 `api/*.go` 协议目录。
- OpenAPI `/api.json` 和 Swagger `/swagger/`。
- 统一响应 `code / message / data`。
- 登录、短信二验、权限和核心后台业务流。
- 商品、第三方对接、开放订单和后台订单主流程。
- 充值风控规则、权限、OpenAPI 暴露和风控记录列表。

契约测试会启动测试态应用，使用 MySQL 测试库 `admin_test`、`miniredis` 和 mock 短信 sender。
