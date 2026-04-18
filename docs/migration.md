# 迁移说明

## 背景

这个仓库原本承载的是一套更分散、更手写的后台实现。
当前迁移工作的目标，是把后端能力收口到仓库根目录，形成统一的 GoFrame 主应用结构，并把目录职责、运行时依赖和测试入口固定下来。

## 已经完成的迁移结果

### 1. 主入口迁到仓库根

当前唯一主入口是：

- `main.go`
- `internal/cmd`
- `internal/bootstrap`

历史 `admin/` 子工程不再是运行入口。

### 2. 协议目录拍平

当前对外协议已经统一到根目录 `api/*.go`：

- `api/auth.go`
- `api/brand.go`
- `api/common.go`
- `api/group.go`
- `api/industry.go`
- `api/log.go`
- `api/product_goods.go`
- `api/product_goods_channel.go`
- `api/product_template.go`
- `api/purchase_limit.go`
- `api/settings.go`
- `api/settings_sms.go`
- `api/settings_system.go`
- `api/subject.go`
- `api/supplier_platform.go`
- `api/user.go`

当前测试也明确约束不能再回到历史 历史嵌套协议包路径。

### 3. 运行时能力收口

运行时核心能力当前统一收口在：

- `internal/app`
- `internal/library/auth`
- `internal/library/sms`
- `internal/library/audit`
- `internal/library/region`

### 4. 后台业务域完成拆分

当前业务实现已经按模块拆分到 `internal/logic/admin`：

- `auth.go`
- `brand.go`
- `industry.go`
- `user.go`
- `group.go`
- `subject.go`
- `config.go`
- `log.go`
- `product_template.go`
- `purchase_limit.go`
- `product_goods_channel_*.go`
- `supplier_platform.go`
- `supplier_platform_balance.go`

### 5. 商品渠道绑定最小闭环已落地

当前商品域已经补齐一组独立的渠道绑定能力：

- 协议拆到 `api/product_goods_channel.go`
- controller 拆到 `internal/controller/admin/product_goods_channel.go`
- logic 按查询 / 写入 / 校验 / 价格拆到 `internal/logic/admin/product_goods_channel*.go`
- MySQL 初始化 SQL 新增 `manifest/sql/006_product_goods_channel_binding.sql`
- 应用启动自建表同步落在 `internal/app/schema.go`

这一轮只覆盖商品列表渠道摘要、绑定弹窗管理和单条自动改价，不包含真实下单、价格通知和批量能力。

## 当前保留的兼容面

迁移后的代码仍然保留了这些兼容约束：

- 接口前缀仍为 `/api/admin/*`
- 响应壳仍为 `code / message / data`
- MySQL、Redis、菜单权限、短信配置、日志表等业务语义继续沿用

## 当前文档应如何理解

这份迁移说明是背景文档，不是日常开发的首要入口。

日常开发更建议优先查看：

- `README.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/testing.md`

## 后续可能继续收尾的方向

- 继续按真实表结构收紧 DAO 自动生成产物
- 继续扩展 `test/integration` 的真实依赖覆盖范围
- 持续清理与当前实现无关的历史描述
