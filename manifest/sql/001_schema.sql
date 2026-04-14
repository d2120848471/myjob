CREATE TABLE IF NOT EXISTS admin_user (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  real_name VARCHAR(64) NOT NULL DEFAULT '',
  phone VARCHAR(20) NOT NULL DEFAULT '',
  group_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  balance_notify TINYINT NOT NULL DEFAULT 0,
  is_business TINYINT NOT NULL DEFAULT 0,
  is_deleted TINYINT NOT NULL DEFAULT 0,
  last_login_ip VARCHAR(45) NULL,
  last_login_at DATETIME NULL,
  token_version INT NOT NULL DEFAULT 0,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_group (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE,
  description VARCHAR(255) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_menu (
  id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL UNIQUE,
  menu_type VARCHAR(16) NOT NULL DEFAULT 'permission',
  menu_level TINYINT NOT NULL DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1,
  super_only TINYINT NOT NULL DEFAULT 0,
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_group_menu (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT UNSIGNED NOT NULL,
  menu_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_group_menu (group_id, menu_id)
);

CREATE TABLE IF NOT EXISTS admin_operation_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NOT NULL,
  admin_name VARCHAR(64) NOT NULL,
  description TEXT NOT NULL,
  ip VARCHAR(45) NOT NULL,
  ip_region VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_login_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  admin_id BIGINT UNSIGNED NOT NULL,
  admin_name VARCHAR(64) NOT NULL,
  ip VARCHAR(45) NOT NULL,
  ip_region VARCHAR(128) NOT NULL,
  created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS admin_subject (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(64) NOT NULL UNIQUE,
  has_tax TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS product_template (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  template_type VARCHAR(32) NOT NULL DEFAULT 'local',
  is_shared TINYINT NOT NULL DEFAULT 0,
  account_name VARCHAR(100) NOT NULL DEFAULT '',
  validate_type INT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_product_template_type_share (template_type, is_shared, id)
);

CREATE TABLE IF NOT EXISTS product_purchase_limit_strategy (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  limit_type TINYINT NOT NULL DEFAULT 1,
  period_type TINYINT NOT NULL DEFAULT 1,
  period INT NOT NULL DEFAULT 1,
  limit_nums INT NOT NULL DEFAULT 0,
  limit_times INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  KEY idx_product_purchase_limit_strategy_keyword (name, id)
);

CREATE TABLE IF NOT EXISTS product_goods (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  goods_code VARCHAR(32) NOT NULL,
  brand_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(255) NOT NULL,
  goods_type VARCHAR(32) NOT NULL,
  supply_type VARCHAR(32) NOT NULL DEFAULT 'channel',
  is_export TINYINT NOT NULL DEFAULT 1,
  is_douyin TINYINT NOT NULL DEFAULT 0,
  has_tax TINYINT NOT NULL DEFAULT 0,
  exception_notify TINYINT NOT NULL DEFAULT 1,
  product_template_id BIGINT UNSIGNED NULL,
  purchase_limit_strategy_id BIGINT UNSIGNED NULL,
  purchase_notice TEXT NULL,
  terminal_price_limit DECIMAL(10,4) NULL,
  balance_limit DECIMAL(10,4) NOT NULL DEFAULT 0.0000,
  default_sell_price DECIMAL(10,4) NULL,
  min_purchase_qty INT NOT NULL DEFAULT 1,
  max_purchase_qty INT NOT NULL DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1,
  is_deleted TINYINT NOT NULL DEFAULT 0,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_product_goods_code (goods_code),
  KEY idx_product_goods_brand (brand_id, is_deleted, id),
  KEY idx_product_goods_status (status, is_deleted, id),
  KEY idx_product_goods_type (goods_type, is_deleted, id),
  KEY idx_product_goods_name (name, is_deleted)
);

CREATE TABLE IF NOT EXISTS system_config (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key VARCHAR(64) NOT NULL UNIQUE,
  config_value TEXT NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
