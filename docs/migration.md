# 迁移说明

## 背景

这个仓库原本承载的是一套更分散、更手写的后台实现。当前迁移工作的目标，是把后端能力收口到仓库根目录，形成统一的 GoFrame 主应用结构，并固定清晰的目录职责、运行时依赖和测试入口。

## 已完成方向

### 主入口收口

当前唯一主入口位于仓库根：

- `main.go`
- `internal/cmd`
- `internal/bootstrap`

历史 `admin/` 子工程不再作为运行入口。

### 协议目录拍平

对外协议统一收口到根目录 `api/*.go`。协议目录保持扁平结构，不再回到历史嵌套协议包路径。

当前协议文件清单和职责以 `module-map.md` 为准。

### 运行时能力收口

运行时核心能力收口在：

- `internal/app`
- `internal/library/auth`
- `internal/library/sms`
- `internal/library/audit`
- `internal/library/region`
- `internal/library/supplierplatform/provider`

### 业务域按职责拆分

后台业务实现按同 package 多文件拆分，典型模式包括：

- `brand*.go`
- `industry*.go`
- `user*.go`
- `product_template*.go`
- `purchase_limit*.go`
- `product_goods*.go`
- `product_goods_channel*.go`
- `supplier_platform*.go`

订单履约单独位于 `internal/logic/order`，不放进商品、第三方平台或通用 helper 文件。

## 当前保留的兼容面

- 后台接口前缀仍为 `/api/admin/*`。
- 开放订单接口前缀为 `/api/open/*`。
- 响应壳仍为 `code / message / data`。
- MySQL、Redis、菜单权限、短信配置、系统参数配置和日志表等业务语义继续沿用。

## 当前文档分工

- 当前功能、路由、权限和文件归属：`module-map.md`。
- 启动链路和分层职责：`architecture.md`。
- 本地开发命令：`development.md`。
- 测试分层和 CI/lint：`testing.md`。

## 后续收尾方向

- 持续清理与当前实现无关的历史描述。
- 持续收紧文档和真实目录、路由、测试、CI 的一致性。
- 在新增重要业务流时同步更新对应专题文档。
