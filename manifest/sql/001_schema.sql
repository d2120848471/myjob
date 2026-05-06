SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS admin_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '员工ID',
  username VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名',
  password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
  real_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '姓名',
  phone VARCHAR(20) NOT NULL DEFAULT '' COMMENT '手机号',
  group_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户组ID',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  balance_notify TINYINT NOT NULL DEFAULT 0 COMMENT '余额通知开关',
  is_business TINYINT NOT NULL DEFAULT 0 COMMENT '是否商务',
  is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
  last_login_ip VARCHAR(45) NULL COMMENT '最后登录IP',
  last_login_at DATETIME NULL COMMENT '最后登录时间',
  token_version INT NOT NULL DEFAULT 0 COMMENT '令牌版本',
  deleted_at DATETIME NULL COMMENT '删除时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='后台员工表';

CREATE TABLE IF NOT EXISTS admin_group (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '用户组ID',
  name VARCHAR(64) NOT NULL UNIQUE COMMENT '用户组名称',
  description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户组描述',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='后台用户组表';

CREATE TABLE IF NOT EXISTS admin_menu (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '菜单ID',
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父菜单ID',
  name VARCHAR(64) NOT NULL COMMENT '菜单名称',
  code VARCHAR(64) NOT NULL UNIQUE COMMENT '权限编码',
  menu_type VARCHAR(16) NOT NULL DEFAULT 'permission' COMMENT '菜单类型',
  menu_level TINYINT NOT NULL DEFAULT 1 COMMENT '菜单层级',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  super_only TINYINT NOT NULL DEFAULT 0 COMMENT '是否仅超级管理员可见',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='后台权限菜单表';

CREATE TABLE IF NOT EXISTS admin_group_menu (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '关联ID',
  group_id BIGINT UNSIGNED NOT NULL COMMENT '用户组ID',
  menu_id BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  UNIQUE KEY uk_group_menu (group_id, menu_id)
) COMMENT='用户组菜单关联表';

CREATE TABLE IF NOT EXISTS admin_operation_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
  admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  admin_name VARCHAR(64) NOT NULL COMMENT '管理员名称',
  description TEXT NOT NULL COMMENT '操作描述',
  ip VARCHAR(45) NOT NULL COMMENT '操作IP',
  ip_region VARCHAR(128) NOT NULL COMMENT 'IP归属地',
  created_at DATETIME NOT NULL COMMENT '创建时间'
) COMMENT='后台操作日志表';

CREATE TABLE IF NOT EXISTS admin_login_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
  admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  admin_name VARCHAR(64) NOT NULL COMMENT '管理员名称',
  ip VARCHAR(45) NOT NULL COMMENT '登录IP',
  ip_region VARCHAR(128) NOT NULL COMMENT 'IP归属地',
  created_at DATETIME NOT NULL COMMENT '创建时间'
) COMMENT='后台登录日志表';

CREATE TABLE IF NOT EXISTS admin_subject (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '主体ID',
  name VARCHAR(64) NOT NULL UNIQUE COMMENT '主体名称',
  has_tax TINYINT NOT NULL DEFAULT 0 COMMENT '是否含税',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='主体配置表';

CREATE TABLE IF NOT EXISTS product_brand (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '品牌ID',
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父品牌ID',
  name VARCHAR(100) NOT NULL COMMENT '品牌名称',
  icon VARCHAR(500) NOT NULL DEFAULT '' COMMENT '品牌图标',
  credential_image VARCHAR(500) NOT NULL DEFAULT '' COMMENT '资质图片',
  description TEXT NULL COMMENT '品牌描述',
  is_visible TINYINT NOT NULL DEFAULT 1 COMMENT '显示状态',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  goods_count INT NOT NULL DEFAULT 0 COMMENT '商品数量',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_product_brand_parent_name (parent_id, name),
  KEY idx_product_brand_parent_sort (parent_id, sort, id)
) COMMENT='商品品牌表';

CREATE TABLE IF NOT EXISTS product_industry (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '行业ID',
  name VARCHAR(100) NOT NULL COMMENT '行业名称',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  brand_count INT NOT NULL DEFAULT 0 COMMENT '品牌数量',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_product_industry_name (name),
  KEY idx_product_industry_sort (sort, id)
) COMMENT='商品行业表';

CREATE TABLE IF NOT EXISTS product_industry_brand (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '关联ID',
  industry_id BIGINT UNSIGNED NOT NULL COMMENT '行业ID',
  brand_id BIGINT UNSIGNED NOT NULL COMMENT '品牌ID',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  UNIQUE KEY uk_product_industry_brand (industry_id, brand_id),
  KEY idx_product_industry_brand_sort (industry_id, sort, id),
  KEY idx_product_industry_brand_brand (brand_id)
) COMMENT='行业品牌关联表';

CREATE TABLE IF NOT EXISTS product_template (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '模板ID',
  title VARCHAR(100) NOT NULL COMMENT '模板标题',
  template_type VARCHAR(32) NOT NULL DEFAULT 'local' COMMENT '模板类型',
  is_shared TINYINT NOT NULL DEFAULT 0 COMMENT '是否共享',
  account_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '模板账号名称',
  validate_type INT NOT NULL DEFAULT 1 COMMENT '校验方式',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  KEY idx_product_template_type_share (template_type, is_shared, id)
) COMMENT='商品模板表';

CREATE TABLE IF NOT EXISTS product_purchase_limit_strategy (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '策略ID',
  name VARCHAR(100) NOT NULL COMMENT '策略名称',
  limit_type TINYINT NOT NULL DEFAULT 1 COMMENT '限制类型',
  period_type TINYINT NOT NULL DEFAULT 1 COMMENT '周期类型',
  period INT NOT NULL DEFAULT 1 COMMENT '周期值',
  limit_nums INT NOT NULL DEFAULT 0 COMMENT '限制数量',
  limit_times INT NOT NULL DEFAULT 0 COMMENT '限制次数',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  KEY idx_product_purchase_limit_strategy_keyword (name, id)
) COMMENT='商品限购策略表';

CREATE TABLE IF NOT EXISTS product_goods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '商品ID',
  goods_code VARCHAR(32) NOT NULL COMMENT '商品编码',
  brand_id BIGINT UNSIGNED NOT NULL COMMENT '品牌ID',
  name VARCHAR(255) NOT NULL COMMENT '商品名称',
  goods_type VARCHAR(32) NOT NULL COMMENT '商品类型',
  supply_type VARCHAR(32) NOT NULL DEFAULT 'channel' COMMENT '供货方式',
  is_export TINYINT NOT NULL DEFAULT 1 COMMENT '是否可导出',
  is_douyin TINYINT NOT NULL DEFAULT 0 COMMENT '是否可抖音',
  has_tax TINYINT NOT NULL DEFAULT 0 COMMENT '是否含税',
  subject_id BIGINT UNSIGNED NULL COMMENT '主体ID',
  exception_notify TINYINT NOT NULL DEFAULT 1 COMMENT '异常提醒开关',
  product_template_id BIGINT UNSIGNED NULL COMMENT '商品模板ID',
  purchase_limit_strategy_id BIGINT UNSIGNED NULL COMMENT '限购策略ID',
  purchase_notice TEXT NULL COMMENT '购买须知',
  terminal_price_limit DECIMAL(10,4) NULL COMMENT '终端限价',
  balance_limit DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '余额限制',
  default_sell_price DECIMAL(10,4) NULL COMMENT '默认售价',
  min_purchase_qty INT NOT NULL DEFAULT 1 COMMENT '最小购买数量',
  max_purchase_qty INT NOT NULL DEFAULT 1 COMMENT '最大购买数量',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
  deleted_at DATETIME NULL COMMENT '删除时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_product_goods_code (goods_code),
  KEY idx_product_goods_brand (brand_id, is_deleted, id),
  KEY idx_product_goods_status (status, is_deleted, id),
  KEY idx_product_goods_type (goods_type, is_deleted, id),
  KEY idx_product_goods_name (name, is_deleted)
) COMMENT='商品表';

CREATE TABLE IF NOT EXISTS system_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
  config_key VARCHAR(64) NOT NULL UNIQUE COMMENT '配置键',
  config_value TEXT NOT NULL COMMENT '配置值',
  description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置描述',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='系统配置表';
