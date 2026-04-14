package app

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS admin_user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    real_name TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    group_id INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    balance_notify INTEGER NOT NULL DEFAULT 0,
    is_business INTEGER NOT NULL DEFAULT 0,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    last_login_ip TEXT,
    last_login_at DATETIME,
    token_version INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS admin_group (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS admin_menu (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    menu_type TEXT NOT NULL DEFAULT 'permission',
    menu_level INTEGER NOT NULL DEFAULT 1,
    status INTEGER NOT NULL DEFAULT 1,
    super_only INTEGER NOT NULL DEFAULT 0,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS admin_group_menu (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    menu_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE(group_id, menu_id)
);
CREATE TABLE IF NOT EXISTS admin_operation_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id INTEGER NOT NULL,
    admin_name TEXT NOT NULL,
    description TEXT NOT NULL,
    ip TEXT NOT NULL,
    ip_region TEXT NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS admin_login_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    admin_id INTEGER NOT NULL,
    admin_name TEXT NOT NULL,
    ip TEXT NOT NULL,
    ip_region TEXT NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS admin_subject (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    has_tax INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS product_brand (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    credential_image TEXT NOT NULL DEFAULT '',
    description TEXT,
    is_visible INTEGER NOT NULL DEFAULT 1,
    sort INTEGER NOT NULL DEFAULT 0,
    goods_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(parent_id, name)
);
CREATE INDEX IF NOT EXISTS idx_product_brand_parent_sort
    ON product_brand(parent_id, sort, id);
CREATE TABLE IF NOT EXISTS product_industry (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    sort INTEGER NOT NULL DEFAULT 0,
    brand_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_product_industry_sort
    ON product_industry(sort, id);
CREATE TABLE IF NOT EXISTS product_industry_brand (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    industry_id INTEGER NOT NULL,
    brand_id INTEGER NOT NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    UNIQUE(industry_id, brand_id)
);
CREATE INDEX IF NOT EXISTS idx_product_industry_brand_sort
    ON product_industry_brand(industry_id, sort, id);
CREATE INDEX IF NOT EXISTS idx_product_industry_brand_brand
    ON product_industry_brand(brand_id);
CREATE TABLE IF NOT EXISTS product_template (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    template_type TEXT NOT NULL DEFAULT 'local',
    is_shared INTEGER NOT NULL DEFAULT 0,
    account_name TEXT NOT NULL DEFAULT '',
    validate_type INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_product_template_type_share
    ON product_template(template_type, is_shared, id);
CREATE TABLE IF NOT EXISTS product_purchase_limit_strategy (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    limit_type INTEGER NOT NULL DEFAULT 1,
    period_type INTEGER NOT NULL DEFAULT 1,
    period INTEGER NOT NULL DEFAULT 1,
    limit_nums INTEGER NOT NULL DEFAULT 0,
    limit_times INTEGER NOT NULL DEFAULT 0,
    status INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_product_purchase_limit_strategy_keyword
    ON product_purchase_limit_strategy(name, id);
CREATE TABLE IF NOT EXISTS product_goods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    goods_code TEXT NOT NULL UNIQUE,
    brand_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    goods_type TEXT NOT NULL,
    supply_type TEXT NOT NULL DEFAULT 'channel',
    is_export INTEGER NOT NULL DEFAULT 1,
    is_douyin INTEGER NOT NULL DEFAULT 0,
    has_tax INTEGER NOT NULL DEFAULT 0,
    exception_notify INTEGER NOT NULL DEFAULT 1,
    product_template_id INTEGER NULL,
    purchase_limit_strategy_id INTEGER NULL,
    purchase_notice TEXT NULL,
    terminal_price_limit TEXT NULL,
    balance_limit TEXT NOT NULL DEFAULT '0.0000',
    default_sell_price TEXT NULL,
    min_purchase_qty INTEGER NOT NULL DEFAULT 1,
    max_purchase_qty INTEGER NOT NULL DEFAULT 1,
    status INTEGER NOT NULL DEFAULT 1,
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_product_goods_brand
    ON product_goods(brand_id, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_product_goods_status
    ON product_goods(status, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_product_goods_type
    ON product_goods(goods_type, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_product_goods_name
    ON product_goods(name, is_deleted);
CREATE TABLE IF NOT EXISTS supplier_platform_type (
    id INTEGER PRIMARY KEY,
    type_name TEXT NOT NULL,
    default_provider_code TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS supplier_platform_account (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    provider_code TEXT NOT NULL,
    provider_name TEXT NOT NULL,
    type_id INTEGER NOT NULL,
    subject_id INTEGER NOT NULL,
    has_tax INTEGER NOT NULL DEFAULT 0,
    domain TEXT NOT NULL,
    backup_domain TEXT NOT NULL DEFAULT '',
    token_id TEXT NOT NULL,
    secret_key TEXT NOT NULL,
    extra_config TEXT NOT NULL DEFAULT '{}',
    threshold_amount TEXT NOT NULL DEFAULT '0.0000',
    sort INTEGER NOT NULL DEFAULT 0,
    crowd_name TEXT NOT NULL DEFAULT '',
    last_balance TEXT NULL,
    last_balance_status INTEGER NOT NULL DEFAULT 0,
    last_balance_message TEXT NOT NULL DEFAULT '',
    last_balance_at DATETIME NULL,
    last_balance_trace_id TEXT NOT NULL DEFAULT '',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(provider_code, subject_id, has_tax, token_id, is_deleted)
);
CREATE INDEX IF NOT EXISTS idx_supplier_platform_filter
    ON supplier_platform_account(type_id, subject_id, has_tax, last_balance_status, is_deleted, sort, id);
CREATE INDEX IF NOT EXISTS idx_supplier_platform_name
    ON supplier_platform_account(name, is_deleted);
CREATE TABLE IF NOT EXISTS supplier_platform_balance_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    platform_id INTEGER NOT NULL,
    operator_id INTEGER NOT NULL DEFAULT 0,
    operator_name TEXT NOT NULL DEFAULT '',
    provider_code TEXT NOT NULL,
    request_url TEXT NOT NULL,
    request_method TEXT NOT NULL DEFAULT 'POST',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    http_status INTEGER NOT NULL DEFAULT 0,
    success INTEGER NOT NULL DEFAULT 0,
    balance_amount TEXT NULL,
    message TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    trace_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_supplier_platform_balance_log_platform
    ON supplier_platform_balance_log(platform_id, created_at);
CREATE INDEX IF NOT EXISTS idx_supplier_platform_balance_log_trace
    ON supplier_platform_balance_log(trace_id);
CREATE TABLE IF NOT EXISTS system_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key TEXT NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)
`

const mysqlSchema = `
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
    updated_at DATETIME NOT NULL,
    KEY idx_admin_user_group (group_id),
    KEY idx_admin_user_deleted (is_deleted)
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
    created_at DATETIME NOT NULL,
    KEY idx_admin_operation_log_created_at (created_at)
);
CREATE TABLE IF NOT EXISTS admin_login_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    admin_id BIGINT UNSIGNED NOT NULL,
    admin_name VARCHAR(64) NOT NULL,
    ip VARCHAR(45) NOT NULL,
    ip_region VARCHAR(128) NOT NULL,
    created_at DATETIME NOT NULL,
    KEY idx_admin_login_log_created_at (created_at)
);
CREATE TABLE IF NOT EXISTS admin_subject (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    has_tax TINYINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS product_brand (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(500) NOT NULL DEFAULT '',
    credential_image VARCHAR(500) NOT NULL DEFAULT '',
    description TEXT NULL,
    is_visible TINYINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    goods_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_product_brand_parent_name (parent_id, name),
    KEY idx_product_brand_parent_sort (parent_id, sort, id)
);
CREATE TABLE IF NOT EXISTS product_industry (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    brand_count INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_product_industry_name (name),
    KEY idx_product_industry_sort (sort, id)
);
CREATE TABLE IF NOT EXISTS product_industry_brand (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    industry_id BIGINT UNSIGNED NOT NULL,
    brand_id BIGINT UNSIGNED NOT NULL,
    sort INT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_product_industry_brand (industry_id, brand_id),
    KEY idx_product_industry_brand_sort (industry_id, sort, id),
    KEY idx_product_industry_brand_brand (brand_id)
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
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)
`
