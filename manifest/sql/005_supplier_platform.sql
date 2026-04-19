SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS supplier_platform_type (
  id INT NOT NULL PRIMARY KEY COMMENT '平台类型ID',
  type_name VARCHAR(64) NOT NULL COMMENT '平台类型名称',
  default_provider_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '默认适配器编码',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '状态',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='第三方平台类型表';

CREATE TABLE IF NOT EXISTS supplier_platform_account (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '平台账号ID',
  name VARCHAR(128) NOT NULL COMMENT '平台名称',
  provider_code VARCHAR(32) NOT NULL COMMENT '适配器编码',
  provider_name VARCHAR(64) NOT NULL COMMENT '适配器名称',
  type_id INT NOT NULL COMMENT '平台类型ID',
  subject_id BIGINT UNSIGNED NOT NULL COMMENT '主体ID',
  has_tax TINYINT NOT NULL DEFAULT 0 COMMENT '是否含税',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '平台业务状态',
  domain VARCHAR(255) NOT NULL COMMENT '主域名',
  backup_domain VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备用域名',
  token_id VARCHAR(128) NOT NULL COMMENT '平台账号ID',
  secret_key VARCHAR(255) NOT NULL COMMENT '平台密钥',
  extra_config TEXT NOT NULL COMMENT '扩展配置',
  threshold_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '余额阈值',
  sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
  crowd_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '群名备注',
  last_balance DECIMAL(18,4) NULL COMMENT '最近余额',
  last_balance_status TINYINT NOT NULL DEFAULT 0 COMMENT '最近余额状态',
  last_balance_message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近余额说明',
  last_balance_at DATETIME NULL COMMENT '最近余额刷新时间',
  last_balance_trace_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最近刷新链路ID',
  is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
  deleted_at DATETIME NULL COMMENT '删除时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_supplier_platform_account_active (provider_code, subject_id, has_tax, token_id, is_deleted),
  KEY idx_supplier_platform_filter (type_id, subject_id, has_tax, last_balance_status, is_deleted, sort, id),
  KEY idx_supplier_platform_name (name, is_deleted)
) COMMENT='第三方平台账号表';

CREATE TABLE IF NOT EXISTS supplier_platform_balance_log (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
  platform_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  operator_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作人名称',
  provider_code VARCHAR(32) NOT NULL COMMENT '适配器编码',
  request_url VARCHAR(512) NOT NULL COMMENT '请求地址',
  request_method VARCHAR(16) NOT NULL DEFAULT 'POST' COMMENT '请求方法',
  request_snapshot TEXT NOT NULL COMMENT '请求快照',
  response_snapshot TEXT NOT NULL COMMENT '响应快照',
  http_status INT NOT NULL DEFAULT 0 COMMENT 'HTTP状态码',
  success TINYINT NOT NULL DEFAULT 0 COMMENT '是否成功',
  balance_amount DECIMAL(18,4) NULL COMMENT '余额值',
  message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '结果说明',
  duration_ms INT NOT NULL DEFAULT 0 COMMENT '耗时毫秒',
  trace_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '链路追踪ID',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  KEY idx_supplier_platform_balance_log_platform (platform_id, created_at),
  KEY idx_supplier_platform_balance_log_trace (trace_id)
) COMMENT='平台余额刷新日志表';

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
