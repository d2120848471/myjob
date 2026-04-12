# OpenAPI B 方案设计说明

## 目标

在当前 GoFrame 后台项目中，按框架原生 OpenAPI 路线接入对外接口文档能力，提供本地开发与测试环境可访问的接口文档页面与规范文件，同时不影响现有 `/api/admin/*` 接口的运行时返回结构。

## 硬约束

1. 现有接口路径、HTTP 方法保持不变。
2. 现有响应包裹结构保持不变，继续输出 `code / msg / data`。
3. 不引入会改变运行时返回字段的默认 GoFrame 响应中间件。
4. 文档入口只在本地开发环境和测试环境开放，不在生产环境开放。
5. 现有前端联调契约以现有 contract tests 为准，改造后必须保持兼容。

## 方案选择

选择 B：按 GoFrame 原生 OpenAPI 路线改造接口声明方式。

不采用 A 的原因：虽然更少改动，但文档维护依赖额外注释体系，不是当前 GoFrame 项目的原生路线。

不采用 C 的原因：工程折中明显，不符合“正规方案”的目标。

## 总体设计

### 1. 路由与 handler 形态调整

把当前部分或全部 `BindHandler("METHOD:/path", func(r *ghttp.Request))` 的写法，逐步转换为 GoFrame 原生可生成 OpenAPI 的函数签名：

`func(ctx context.Context, req *Req) (res *Res, err error)`

请求结构与响应结构显式定义，并在请求结构上补充 `g.Meta` 元信息，包括：

- `path`
- `method`
- `tags`
- `summary`
- 必要时补充 `description`
- 鉴权相关 `security`

### 2. 保留现有返回结构

运行时返回不使用 GoFrame 默认的 `MiddlewareHandlerResponse` 作为最终对外协议实现方式，避免把现有 `msg` 改成 `message`。

实际接口仍然统一输出：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

文档层通过 OpenAPI 公共响应壳声明，表达统一的 `code / msg / data` 结构。

### 3. 环境开关

文档能力只在以下环境启用：

- 本地开发环境
- 测试环境

生产环境不暴露：

- OpenAPI JSON
- 文档页面

### 4. 文档入口

计划提供两个入口：

- `/docs/api/openapi.json`
- `/docs/api/`

其中 `/docs/api/` 展示文档页面，`/docs/api/openapi.json` 提供规范文件。

### 5. 响应建模

为避免文档与实际响应不一致，需要在 OpenAPI 配置中显式定义公共响应结构，数据体由各接口业务响应结构替换到 `data` 字段。

## 分阶段实施

### Phase 1

先完成基础能力搭建：

- 文档开关
- OpenAPI 路径
- 文档页面路径
- 公共响应壳建模
- 选取一组核心接口做首批原生化改造

建议首批接口：

- 登录
- 短信发送
- 短信校验
- me
- logout

原因：这组接口前端最先用到，也最适合作为鉴权和公共响应壳的验证样板。

### Phase 2

逐步补齐：

- 用户管理
- 用户组与授权
- 主体管理
- 短信配置
- 日志查询

## 风险与规避

### 风险 1：返回结构漂移

如果误接 GoFrame 默认响应中间件，可能把 `msg` 变成 `message`。

规避：保留当前自定义响应输出层，并以 contract tests 校验。

### 风险 2：接口行为变化

如果把 controller 改写时顺手调整了解析逻辑、错误码或状态码，可能影响前端。

规避：先从 contract tests 列出现有行为，再按测试驱动逐个迁移。

### 风险 3：文档和实际不一致

如果业务返回仍大量使用 `map[string]any`，容易导致文档字段模糊。

规避：迁移时把关键接口响应结构显式化。

## 验证策略

1. 先补或复用 contract tests，确保现有接口返回契约不变。
2. 对改造后的接口运行定向测试。
3. 启动测试应用或本地服务，验证：
   - 文档页可访问
   - openapi.json 可访问
   - 现有接口响应体仍为 `code / msg / data`
4. 最后跑：
   - `go test ./... -count=1 -timeout 60s`
   - `go build ./...`

## 当前结论

该方案是当前项目最正规的实现路径，同时可以通过“文档原生化、响应契约不变”的方式，避免前端联调被破坏。
