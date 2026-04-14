# 商品管理一期设计说明

**目标**

在当前后台仓库内补齐商品管理一期后端闭环，只交付商品 CRUD、表单聚合下拉、菜单权限同步、契约测试，以及品牌/模板/购买数量限制策略的商品反向引用校验。

**范围边界**

- 只做 `product_goods` 主数据，不实现渠道绑定、自动改价、导出逻辑、库存余额联动。
- 商品展示名称全部实时关联品牌、模板、策略主数据，不做名称快照冗余。
- 旧模块仅补删除前引用校验，不把商品逻辑散落回旧模块。

**实现方式**

1. 新增独立商品模块，沿用现有 `api -> controller -> service -> logic` 分层。
2. 新增 `product.goods` 菜单权限、默认组授权和路由挂载，保持运行时与部署 SQL 双份同步。
3. `product_goods` 新增、改品牌、软删时，和 `product_brand.goods_count` 更新放在同一事务。
4. `goods_code` 在插入后按 `GD` + 10 位补零自增 ID 生成，软删后不复用。
5. 商品详情允许回显已禁用但仍被当前商品绑定的旧策略；表单聚合接口仍只返回启用中的策略。

**数据与校验**

- `brand_id` 必须是叶子品牌。
- `supply_type` 本期固定只允许 `channel`。
- `goods_type` 只允许 `card_secret`、`direct_recharge`。
- `product_template_id` 可空，但传值时必须存在。
- `purchase_limit_strategy_id` 可空；创建时必须是启用策略，编辑时仅在“更换成新策略”时要求启用。
- `balance_limit=0` 表示不限制；`min_purchase_qty >= 1`；`max_purchase_qty >= min_purchase_qty`。
- 列表按父级品牌筛选时，需要递归展开所有子孙品牌后再查询。

**测试策略**

- 先补契约测试覆盖 OpenAPI、权限、菜单、CRUD、搜索、软删、表单聚合和反向引用校验。
- 通过契约测试驱动接口和逻辑实现，再跑全量 `go test ./...` 回归。
