CREATE TABLE IF NOT EXISTS supplier_platform_type (
  id INT NOT NULL PRIMARY KEY,
  type_name VARCHAR(64) NOT NULL,
  default_provider_code VARCHAR(32) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  sort INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS supplier_platform_account (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  provider_code VARCHAR(32) NOT NULL,
  provider_name VARCHAR(64) NOT NULL,
  type_id INT NOT NULL,
  subject_id BIGINT UNSIGNED NOT NULL,
  has_tax TINYINT NOT NULL DEFAULT 0,
  domain VARCHAR(255) NOT NULL,
  backup_domain VARCHAR(255) NOT NULL DEFAULT '',
  token_id VARCHAR(128) NOT NULL,
  secret_key VARCHAR(255) NOT NULL,
  extra_config TEXT NOT NULL,
  threshold_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
  sort INT NOT NULL DEFAULT 0,
  crowd_name VARCHAR(128) NOT NULL DEFAULT '',
  last_balance DECIMAL(18,4) NULL,
  last_balance_status TINYINT NOT NULL DEFAULT 0,
  last_balance_message VARCHAR(255) NOT NULL DEFAULT '',
  last_balance_at DATETIME NULL,
  last_balance_trace_id VARCHAR(64) NOT NULL DEFAULT '',
  is_deleted TINYINT NOT NULL DEFAULT 0,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  UNIQUE KEY uk_supplier_platform_account_active (provider_code, subject_id, has_tax, token_id, is_deleted),
  KEY idx_supplier_platform_filter (type_id, subject_id, has_tax, last_balance_status, is_deleted, sort, id),
  KEY idx_supplier_platform_name (name, is_deleted)
);

CREATE TABLE IF NOT EXISTS supplier_platform_balance_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  platform_id BIGINT UNSIGNED NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_name VARCHAR(64) NOT NULL DEFAULT '',
  provider_code VARCHAR(32) NOT NULL,
  request_url VARCHAR(512) NOT NULL,
  request_method VARCHAR(16) NOT NULL DEFAULT 'POST',
  request_snapshot TEXT NOT NULL,
  response_snapshot TEXT NOT NULL,
  http_status INT NOT NULL DEFAULT 0,
  success TINYINT NOT NULL DEFAULT 0,
  balance_amount DECIMAL(18,4) NULL,
  message VARCHAR(255) NOT NULL DEFAULT '',
  duration_ms INT NOT NULL DEFAULT 0,
  trace_id VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  KEY idx_supplier_platform_balance_log_platform (platform_id, created_at),
  KEY idx_supplier_platform_balance_log_trace (trace_id)
);

INSERT INTO supplier_platform_type (
  id, type_name, default_provider_code, status, sort, created_at, updated_at
) VALUES
  (6,  '云发卡', 'kakayun',    1,  6, NOW(), NOW()),
  (7,  '同系统', 'youkayun',   1,  7, NOW(), NOW()),
  (15, '星海',   'xinghai',    1, 15, NOW(), NOW()),
  (35, '星权益', 'xingquanyi', 1, 35, NOW(), NOW()),
  (56, '雅兰芳', 'feisuyuan',  1, 56, NOW(), NOW()),
  (72, '卡速售', 'kasushou',   1, 72, NOW(), NOW()),
  (73, '卡易信', 'kayixin',    1, 73, NOW(), NOW()),
  (81, '聚浪云', 'julangyun',  1, 81, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  type_name = VALUES(type_name),
  default_provider_code = VALUES(default_provider_code),
  status = VALUES(status),
  sort = VALUES(sort),
  updated_at = VALUES(updated_at);
