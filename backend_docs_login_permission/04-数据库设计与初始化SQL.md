# 04｜数据库设计与初始化SQL

## 1. 数据库说明

数据库名：`admin`

本项目数据库对象只包含以下 8 张表：

1. `admin_user`
2. `admin_group`
3. `admin_menu`
4. `admin_group_menu`
5. `admin_operation_log`
6. `admin_login_log`
7. `admin_subject`
8. `system_config`

## 2. 设计原则

- 所有表统一使用 `utf8mb4`
- 所有主键统一使用 `bigint unsigned`
- 软删除只用于 `admin_user`
- 权限菜单固定存于 `admin_menu`
- 用户组与菜单关系通过 `admin_group_menu`
- 日志表只增不删
- 配置表使用 key-value 设计

## 3. 初始化顺序

```text
001_schema.sql       -> 建表
002_seed_menu.sql    -> 初始化权限菜单
003_seed_admin.sql   -> 初始化超级管理员
004_seed_config.sql  -> 初始化短信配置默认值
```

## 4. 特别说明

### 4.1 超级管理员初始化字段

由于需求文档没有给默认手机号和默认密码，**这里不硬编码真实初始凭证**。  
初始化 SQL 中使用占位符：

- `{{SUPER_ADMIN_PHONE}}`
- `{{SUPER_ADMIN_BCRYPT_HASH}}`

交给 Codex 实施时，必须在部署脚本或初始化脚本中替换后执行。

### 4.2 删除用户组的人数校验

为避免“回收站员工恢复时用户组已不存在”的脏数据问题，删除用户组时建议校验**该组下全部员工**，包括回收站中的员工。  
这样可以保证恢复逻辑稳定，不会出现恢复后 `group_id` 指向无效记录。

## 5. 建表 SQL

### 5.1 `admin_user`

```sql
CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(64) NOT NULL COMMENT '登录用户名，活跃用户唯一',
  `password_hash` varchar(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  `real_name` varchar(64) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `phone` varchar(20) NOT NULL DEFAULT '' COMMENT '手机号',
  `group_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '用户组ID，0=超级管理员',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1正常 0禁用',
  `balance_notify` tinyint NOT NULL DEFAULT 0 COMMENT '余额阈值通知：1开启 0关闭',
  `is_business` tinyint NOT NULL DEFAULT 0 COMMENT '是否商务：1是 0否',
  `is_deleted` tinyint NOT NULL DEFAULT 0 COMMENT '是否删除：1删除 0未删除',
  `last_login_ip` varchar(45) DEFAULT NULL COMMENT '上次登录成功IP',
  `last_login_at` datetime DEFAULT NULL COMMENT '上次登录成功时间',
  `token_version` int unsigned NOT NULL DEFAULT 0 COMMENT 'Token 版本号，用于强制失效',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_user_username` (`username`),
  KEY `idx_admin_user_group_id` (`group_id`),
  KEY `idx_admin_user_status_deleted` (`status`, `is_deleted`),
  KEY `idx_admin_user_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员/员工表';
```

### 5.2 `admin_group`

```sql
CREATE TABLE IF NOT EXISTS `admin_group` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) NOT NULL COMMENT '用户组名称',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1正常 0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_group_name` (`name`),
  KEY `idx_admin_group_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组/部门表';
```

### 5.3 `admin_menu`

```sql
CREATE TABLE IF NOT EXISTS `admin_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT '父级ID，0=顶级',
  `name` varchar(128) NOT NULL COMMENT '菜单/权限名称',
  `code` varchar(128) NOT NULL COMMENT '权限码，全局唯一',
  `menu_level` tinyint NOT NULL DEFAULT 1 COMMENT '层级：1模块 2页面 3按钮',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1启用 0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_menu_code` (`code`),
  KEY `idx_admin_menu_parent_id` (`parent_id`),
  KEY `idx_admin_menu_status_sort` (`status`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限菜单定义表';
```

### 5.4 `admin_group_menu`

```sql
CREATE TABLE IF NOT EXISTS `admin_group_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_id` bigint unsigned NOT NULL COMMENT '用户组ID',
  `menu_id` bigint unsigned NOT NULL COMMENT '菜单ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_group_menu` (`group_id`, `menu_id`),
  KEY `idx_admin_group_menu_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组-菜单关联表';
```

### 5.5 `admin_operation_log`

```sql
CREATE TABLE IF NOT EXISTS `admin_operation_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` bigint unsigned NOT NULL COMMENT '操作人ID',
  `admin_name` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人姓名快照',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '业务中文描述',
  `ip` varchar(45) NOT NULL DEFAULT '' COMMENT '操作IP',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_operation_log_admin_id` (`admin_id`),
  KEY `idx_admin_operation_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
```

### 5.6 `admin_login_log`

```sql
CREATE TABLE IF NOT EXISTS `admin_login_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `admin_id` bigint unsigned NOT NULL COMMENT '登录人ID',
  `admin_name` varchar(64) NOT NULL DEFAULT '' COMMENT '登录人姓名快照',
  `ip` varchar(45) NOT NULL DEFAULT '' COMMENT '登录IP',
  `ip_region` varchar(100) NOT NULL DEFAULT '' COMMENT 'IP归属地',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_login_log_admin_id` (`admin_id`),
  KEY `idx_admin_login_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='登录日志表';
```

### 5.7 `admin_subject`

```sql
CREATE TABLE IF NOT EXISTS `admin_subject` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(128) NOT NULL COMMENT '主体名称',
  `has_tax` tinyint NOT NULL DEFAULT 0 COMMENT '是否含税：1是 0否',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_subject_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='主体配置表';
```

### 5.8 `system_config`

```sql
CREATE TABLE IF NOT EXISTS `system_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `config_key` varchar(64) NOT NULL COMMENT '配置键',
  `config_value` text COMMENT '配置值',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '配置说明',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';
```

## 6. 初始化 SQL

### 6.1 `002_seed_menu.sql`

> 注：未在需求文档中明确给出中文名称的页面/按钮，`name` 暂以 `code` 同名写入。
> 后续如补充正式菜单文案，可直接批量更新 `admin_menu.name`。

```sql
TRUNCATE TABLE `admin_group_menu`;
TRUNCATE TABLE `admin_menu`;

INSERT INTO `admin_menu` (`id`, `parent_id`, `name`, `code`, `menu_level`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(1, 0, '首页', 'home.manager', 1, 1, 1, NOW(), NOW()),
(2, 1, '首页', 'home.index', 2, 1, 1, NOW(), NOW()),
(3, 0, '网站设置', 'setting.manage', 1, 2, 1, NOW(), NOW()),
(4, 3, '首页装修设置', 'setting.indexset', 2, 1, 1, NOW(), NOW()),
(5, 3, '系统参数设置', 'setting.general', 2, 2, 1, NOW(), NOW()),
(6, 3, 'setting.news', 'setting.news', 2, 3, 1, NOW(), NOW()),
(7, 3, 'setting.generalpay', 'setting.generalpay', 2, 4, 1, NOW(), NOW()),
(8, 3, 'setting.paymsg', 'setting.paymsg', 2, 5, 1, NOW(), NOW()),
(9, 3, 'setting.cron', 'setting.cron', 2, 6, 1, NOW(), NOW()),
(10, 3, 'setting.ipblack', 'setting.ipblack', 2, 7, 1, NOW(), NOW()),
(11, 3, '超级密码管理', 'setting.supperpassword', 2, 8, 1, NOW(), NOW()),
(12, 3, 'setting.scene', 'setting.scene', 2, 9, 1, NOW(), NOW()),
(13, 0, '机器人模板', 'robot.template', 1, 3, 1, NOW(), NOW()),
(14, 0, '商品管理', 'goods.manage', 1, 4, 1, NOW(), NOW()),
(15, 14, 'goods.categorylist', 'goods.categorylist', 2, 1, 1, NOW(), NOW()),
(16, 14, 'goods.industry', 'goods.industry', 2, 2, 1, NOW(), NOW()),
(17, 14, 'goods.goodslist', 'goods.goodslist', 2, 3, 1, NOW(), NOW()),
(18, 14, 'goods.crossdocking', 'goods.crossdocking', 2, 4, 1, NOW(), NOW()),
(19, 14, 'goods.balance', 'goods.balance', 2, 5, 1, NOW(), NOW()),
(20, 14, 'goods.risking', 'goods.risking', 2, 6, 1, NOW(), NOW()),
(21, 14, 'goods.pricechange', 'goods.pricechange', 2, 7, 1, NOW(), NOW()),
(22, 14, 'goods.send', 'goods.send', 2, 8, 1, NOW(), NOW()),
(23, 14, 'goods.tpllist', 'goods.tpllist', 2, 9, 1, NOW(), NOW()),
(24, 14, 'goods.configtpl', 'goods.configtpl', 2, 10, 1, NOW(), NOW()),
(25, 14, 'goods.ploy', 'goods.ploy', 2, 11, 1, NOW(), NOW()),
(26, 14, 'goods.subscribe', 'goods.subscribe', 2, 12, 1, NOW(), NOW()),
(27, 0, '自动改价', 'auto.record', 1, 5, 1, NOW(), NOW()),
(28, 0, '客户管理', 'customer.manage', 1, 6, 1, NOW(), NOW()),
(29, 28, 'customer.list', 'customer.list', 2, 1, 1, NOW(), NOW()),
(30, 28, 'customer.money', 'customer.money', 2, 2, 1, NOW(), NOW()),
(31, 28, 'customer.lockuser', 'customer.lockuser', 2, 3, 1, NOW(), NOW()),
(32, 28, 'customer.credit', 'customer.credit', 2, 4, 1, NOW(), NOW()),
(33, 28, 'customer.account', 'customer.account', 2, 5, 1, NOW(), NOW()),
(34, 28, 'customer.alipaybind', 'customer.alipaybind', 2, 6, 1, NOW(), NOW()),
(35, 0, '订单管理', 'order.manage', 1, 7, 1, NOW(), NOW()),
(36, 35, '交易记录', 'order.list', 2, 1, 1, NOW(), NOW()),
(37, 36, '导出', 'order.list.export', 3, 1, 1, NOW(), NOW()),
(38, 36, '退款', 'order.list.refund', 3, 2, 1, NOW(), NOW()),
(39, 35, 'order.jdlist', 'order.jdlist', 2, 2, 1, NOW(), NOW()),
(40, 39, '导出', 'order.jdlist.export', 3, 1, 1, NOW(), NOW()),
(41, 35, 'order.callbacklist', 'order.callbacklist', 2, 3, 1, NOW(), NOW()),
(42, 35, 'order.back', 'order.back', 2, 4, 1, NOW(), NOW()),
(43, 35, 'order.batch', 'order.batch', 2, 5, 1, NOW(), NOW()),
(44, 35, 'order.intercept', 'order.intercept', 2, 6, 1, NOW(), NOW()),
(45, 35, 'order.handle', 'order.handle', 2, 7, 1, NOW(), NOW()),
(46, 35, 'order.dcorder', 'order.dcorder', 2, 8, 1, NOW(), NOW()),
(47, 46, '导出', 'order.dclist.export', 3, 1, 1, NOW(), NOW()),
(48, 35, 'order.xslist', 'order.xslist', 2, 9, 1, NOW(), NOW()),
(49, 35, 'order.zglist', 'order.zglist', 2, 10, 1, NOW(), NOW()),
(50, 35, 'order.ylist', 'order.ylist', 2, 11, 1, NOW(), NOW()),
(51, 0, '财务管理', 'financial.manage', 1, 8, 1, NOW(), NOW()),
(52, 51, 'financial.payinfo', 'financial.payinfo', 2, 1, 1, NOW(), NOW()),
(53, 51, 'financial.change', 'financial.change', 2, 2, 1, NOW(), NOW()),
(54, 51, 'financial.account', 'financial.account', 2, 3, 1, NOW(), NOW()),
(55, 51, 'financial.free', 'financial.free', 2, 4, 1, NOW(), NOW()),
(56, 0, '数据分析', 'analysis.manage', 1, 9, 1, NOW(), NOW()),
(57, 56, 'analysis.rank', 'analysis.rank', 2, 1, 1, NOW(), NOW()),
(58, 56, 'analysis.goods-sale', 'analysis.goods-sale', 2, 2, 1, NOW(), NOW()),
(59, 56, 'analysis.member-buy', 'analysis.member-buy', 2, 3, 1, NOW(), NOW()),
(60, 56, 'analysis.card-sale', 'analysis.card-sale', 2, 4, 1, NOW(), NOW()),
(61, 56, 'analysis.docking-order', 'analysis.docking-order', 2, 5, 1, NOW(), NOW()),
(62, 56, 'analysis.docking-goods', 'analysis.docking-goods', 2, 6, 1, NOW(), NOW()),
(63, 56, 'analysis.order-profit', 'analysis.order-profit', 2, 7, 1, NOW(), NOW()),
(64, 56, 'analysis.pay-channel', 'analysis.pay-channel', 2, 8, 1, NOW(), NOW()),
(65, 56, 'analysis.data', 'analysis.data', 2, 9, 1, NOW(), NOW()),
(66, 56, 'analysis.compare', 'analysis.compare', 2, 10, 1, NOW(), NOW()),
(67, 56, 'analysis.sale', 'analysis.sale', 2, 11, 1, NOW(), NOW()),
(68, 56, 'analysis.business', 'analysis.business', 2, 12, 1, NOW(), NOW()),
(69, 0, '员工管理', 'admin.manage', 1, 10, 1, NOW(), NOW()),
(70, 69, '管理员列表', 'admin.list', 2, 1, 1, NOW(), NOW()),
(71, 69, '操作日志', 'admin.action', 2, 2, 1, NOW(), NOW()),
(72, 69, '登录日志', 'admin.loginlog', 2, 3, 1, NOW(), NOW()),
(73, 69, '部门管理', 'admin.department', 2, 4, 1, NOW(), NOW()),
(74, 0, '卡密兑换', 'km.manage', 1, 11, 1, NOW(), NOW()),
(75, 74, 'km.list', 'km.list', 2, 1, 1, NOW(), NOW()),
(76, 74, 'km.kami', 'km.kami', 2, 2, 1, NOW(), NOW()),
(77, 74, 'km.record', 'km.record', 2, 3, 1, NOW(), NOW()),
(78, 0, '消息管理', 'message.manage', 1, 12, 1, NOW(), NOW()),
(79, 78, '消息列表', 'message.index', 2, 1, 1, NOW(), NOW()),
(80, 0, '我的导出', 'export.manage', 1, 13, 1, NOW(), NOW()),
(81, 80, '导出列表', 'export.index', 2, 1, 1, NOW(), NOW()),
(82, 0, '供货商管理', 'supplier.manager', 1, 14, 1, NOW(), NOW()),
(83, 82, 'supplier.index', 'supplier.index', 2, 1, 1, NOW(), NOW()),
(84, 82, 'supplier.goods', 'supplier.goods', 2, 2, 1, NOW(), NOW()),
(85, 82, 'supplier.money', 'supplier.money', 2, 3, 1, NOW(), NOW()),
(86, 82, 'supplier.withdraw', 'supplier.withdraw', 2, 4, 1, NOW(), NOW()),
(87, 82, 'supplier.category', 'supplier.category', 2, 5, 1, NOW(), NOW()),
(88, 82, 'supplier.record', 'supplier.record', 2, 6, 1, NOW(), NOW()),
(89, 82, 'supplier.ban', 'supplier.ban', 2, 7, 1, NOW(), NOW()),
(90, 82, 'supplier.order', 'supplier.order', 2, 8, 1, NOW(), NOW());
```

### 6.2 `003_seed_admin.sql`

```sql
INSERT INTO `admin_user`
(
  `id`,
  `username`,
  `password_hash`,
  `real_name`,
  `phone`,
  `group_id`,
  `status`,
  `balance_notify`,
  `is_business`,
  `is_deleted`,
  `last_login_ip`,
  `last_login_at`,
  `token_version`,
  `deleted_at`,
  `created_at`,
  `updated_at`
)
VALUES
(
  1,
  'admin',
  '{{SUPER_ADMIN_BCRYPT_HASH}}',
  '系统管理员',
  '{{SUPER_ADMIN_PHONE}}',
  0,
  1,
  0,
  1,
  0,
  NULL,
  NULL,
  0,
  NULL,
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  `password_hash` = VALUES(`password_hash`),
  `real_name` = VALUES(`real_name`),
  `phone` = VALUES(`phone`),
  `status` = VALUES(`status`),
  `is_business` = VALUES(`is_business`),
  `is_deleted` = VALUES(`is_deleted`),
  `updated_at` = NOW();
```

### 6.3 `004_seed_config.sql`

```sql
INSERT INTO `system_config` (`config_key`, `config_value`, `description`, `created_at`, `updated_at`) VALUES
('sms_access_key', '', '阿里云 AccessKey ID', NOW(), NOW()),
('sms_access_key_secret', '', '阿里云 AccessKey Secret', NOW(), NOW()),
('sms_sign_name', '', '短信签名', NOW(), NOW()),
('sms_template_code', '', '短信模板编号', NOW(), NOW()),
('sms_expire_minutes', '30', '验证码有效期（分钟）', NOW(), NOW()),
('sms_interval_minutes', '1', '验证码发送间隔（分钟）', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `config_value` = VALUES(`config_value`),
  `description` = VALUES(`description`),
  `updated_at` = NOW();
```

## 7. 索引说明

### 7.1 `admin_user`

- `uk_admin_user_username`：保证当前活跃用户名唯一  
  删除员工时先改用户名再软删除，因此不会与活跃用户冲突
- `idx_admin_user_status_deleted`：登录、中间件鉴权、员工列表常用
- `idx_admin_user_group_id`：用户组人数统计、员工列表关联查询常用

### 7.2 日志表

日志表重点索引：

- `admin_id`
- `created_at`

因为查询主要按人员 + 时间范围分页。  
`description keyword` 模糊查询暂不单独上全文索引，当前数据量足够；后续如果模糊搜索成本升高，再升级全文检索方案。

## 8. 推荐的 SQL 文件拆分

虽然本交付统一输出为 Markdown，但真正落地时建议拆到：

```text
manifest/sql/
├── 001_schema.sql
├── 002_seed_menu.sql
├── 003_seed_admin.sql
└── 004_seed_config.sql
```

## 9. DAO 生成要求

执行顺序：

```bash
# 1) 先执行 manifest/sql 中的建表和 seed
# 2) 再在项目根目录执行
gf gen dao
```

生成产物必须进入：

- `internal/dao`
- `internal/model/do`
- `internal/model/entity`

## 10. 数据库侧验收清单

- 8 张表全部创建成功
- 超管账号 `id=1` 成功写入
- `admin_menu` 90 条菜单/权限数据写入成功
- `system_config` 6 条短信配置初始值写入成功
- `admin_user.username` 唯一约束生效
- `admin_group_menu(group_id, menu_id)` 唯一约束生效
