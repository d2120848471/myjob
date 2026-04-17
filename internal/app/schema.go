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
    subject_id INTEGER NULL,
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
);

CREATE TABLE IF NOT EXISTS product_goods_channel_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    goods_id INTEGER NOT NULL UNIQUE,
    smart_replenish_enabled INTEGER NOT NULL DEFAULT 0,
    attempt_timeout_enabled INTEGER NOT NULL DEFAULT 0,
    attempt_timeout_minutes INTEGER NOT NULL DEFAULT 0,
    route_mode TEXT NOT NULL DEFAULT 'fixed_order',
    sync_cost_enabled INTEGER NOT NULL DEFAULT 0,
    sync_goods_name_enabled INTEGER NOT NULL DEFAULT 0,
    allow_loss INTEGER NOT NULL DEFAULT 0,
    max_loss_amount TEXT NULL,
    is_bundle INTEGER NOT NULL DEFAULT 0,
    min_channel_cost_snapshot TEXT NULL,
    bound_channel_count_snapshot INTEGER NOT NULL DEFAULT 0,
    primary_binding_id INTEGER NULL,
    primary_channel_name_snapshot TEXT NOT NULL DEFAULT '',
    channel_auto_price_status_snapshot INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS product_goods_channel_binding (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    goods_id INTEGER NOT NULL,
    platform_account_id INTEGER NOT NULL,
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    cost_price TEXT NOT NULL DEFAULT '0.0000',
    tax_adjust_direction TEXT NOT NULL DEFAULT 'none',
    tax_adjust_rate TEXT NOT NULL DEFAULT '0.0000',
    tax_adjust_amount TEXT NOT NULL DEFAULT '0.0000',
    dock_status TEXT NOT NULL DEFAULT 'enabled',
    sort INTEGER NOT NULL DEFAULT 0,
    weight INTEGER NOT NULL DEFAULT 0,
    start_time TEXT NOT NULL DEFAULT '',
    end_time TEXT NOT NULL DEFAULT '',
    validate_template_id INTEGER NULL,
    is_auto_change INTEGER NOT NULL DEFAULT 0,
    add_type TEXT NOT NULL DEFAULT 'fixed',
    default_price TEXT NOT NULL DEFAULT '0.0000',
    lock_price TEXT NOT NULL DEFAULT '0.0000',
    symbol_price TEXT NOT NULL DEFAULT '0.0000',
    max_price TEXT NOT NULL DEFAULT '0.0000',
    min_price TEXT NOT NULL DEFAULT '0.0000',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(goods_id, platform_account_id, supplier_goods_no, is_deleted)
);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_goods_id
    ON product_goods_channel_binding(goods_id);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_platform_account_id
    ON product_goods_channel_binding(platform_account_id);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_goods_status
    ON product_goods_channel_binding(goods_id, dock_status, is_deleted);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_goods_sort
    ON product_goods_channel_binding(goods_id, sort, is_deleted);

CREATE TABLE IF NOT EXISTS trade_order (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no TEXT NOT NULL UNIQUE,
    caller_id INTEGER NOT NULL,
    client_order_no TEXT NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code_snapshot TEXT NOT NULL DEFAULT '',
    goods_name_snapshot TEXT NOT NULL DEFAULT '',
    binding_id INTEGER NOT NULL DEFAULT 0,
    platform_account_id INTEGER NOT NULL DEFAULT 0,
    route_mode_snapshot TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL,
    success_quantity INTEGER NOT NULL DEFAULT 0,
    failed_quantity INTEGER NOT NULL DEFAULT 0,
    payload_json TEXT NOT NULL DEFAULT '',
    sale_price TEXT NOT NULL DEFAULT '0.0000',
    total_amount TEXT NOT NULL DEFAULT '0.0000',
    source_cost_price_snapshot TEXT NOT NULL DEFAULT '0.0000',
    cost_price_snapshot TEXT NOT NULL DEFAULT '0.0000',
    tax_adjust_direction TEXT NOT NULL DEFAULT 'none',
    tax_adjust_rate TEXT NOT NULL DEFAULT '0.0000',
    tax_adjust_amount TEXT NOT NULL DEFAULT '0.0000',
    loss_order INTEGER NOT NULL DEFAULT 0,
    loss_amount TEXT NOT NULL DEFAULT '0.0000',
    channel_order_no TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'created',
    failure_reason TEXT NOT NULL DEFAULT '',
    finished_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(caller_id, client_order_no)
);
CREATE INDEX IF NOT EXISTS idx_trade_order_goods_id
    ON trade_order(goods_id);
CREATE INDEX IF NOT EXISTS idx_trade_order_status
    ON trade_order(status);
CREATE INDEX IF NOT EXISTS idx_trade_order_created_at
    ON trade_order(created_at);

CREATE TABLE IF NOT EXISTS trade_order_attempt (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    binding_id INTEGER NOT NULL,
    platform_account_id INTEGER NOT NULL,
    provider_code TEXT NOT NULL,
    fulfillment_no TEXT NOT NULL,
    attempt_quantity INTEGER NOT NULL,
    attempt_no INTEGER NOT NULL,
    provider_request_order_no TEXT NOT NULL UNIQUE,
    channel_order_no TEXT NOT NULL DEFAULT '',
    attempt_status TEXT NOT NULL DEFAULT 'created',
    upstream_status TEXT NOT NULL DEFAULT '',
    binding_channel_name_snapshot TEXT NOT NULL DEFAULT '',
    binding_supplier_goods_no_snapshot TEXT NOT NULL DEFAULT '',
    source_cost_price_snapshot TEXT NOT NULL DEFAULT '0.0000',
    cost_price_snapshot TEXT NOT NULL DEFAULT '0.0000',
    sale_price_snapshot TEXT NOT NULL DEFAULT '0.0000',
    loss_amount_snapshot TEXT NOT NULL DEFAULT '0.0000',
    request_url TEXT NOT NULL DEFAULT '',
    request_method TEXT NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL DEFAULT '',
    request_payload TEXT NOT NULL DEFAULT '',
    response_payload TEXT NOT NULL DEFAULT '',
    http_status INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    error_category TEXT NOT NULL DEFAULT '',
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    query_count INTEGER NOT NULL DEFAULT 0,
    last_query_at DATETIME NULL,
    next_query_at DATETIME NULL,
    query_deadline_at DATETIME NULL,
    callback_payload TEXT NOT NULL DEFAULT '',
    callback_received_at DATETIME NULL,
    callback_processed_at DATETIME NULL,
    trace_id TEXT NOT NULL DEFAULT '',
    finished_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(order_id, fulfillment_no, attempt_no)
);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_order_id
    ON trade_order_attempt(order_id);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_order_fulfillment
    ON trade_order_attempt(order_id, fulfillment_no);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_binding_id
    ON trade_order_attempt(binding_id);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_channel_order_no
    ON trade_order_attempt(channel_order_no);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_attempt_status
    ON trade_order_attempt(attempt_status);
CREATE INDEX IF NOT EXISTS idx_trade_order_attempt_next_query_at
    ON trade_order_attempt(next_query_at);

CREATE TABLE IF NOT EXISTS provider_callback_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL,
    provider_request_order_no TEXT NOT NULL DEFAULT '',
    channel_order_no TEXT NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL,
    request_body TEXT NOT NULL,
    verify_result TEXT NOT NULL DEFAULT '',
    process_result TEXT NOT NULL DEFAULT '',
    ack_body TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    UNIQUE(provider_code, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_provider_callback_log_provider_request_order_no
    ON provider_callback_log(provider_request_order_no);
CREATE INDEX IF NOT EXISTS idx_provider_callback_log_channel_order_no
    ON provider_callback_log(channel_order_no);

CREATE TABLE IF NOT EXISTS provider_price_notify_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL DEFAULT 0,
    idempotency_key TEXT NOT NULL,
    supplier_goods_no TEXT NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL,
    request_body TEXT NOT NULL,
    source_cost_price_new TEXT NOT NULL DEFAULT '0.0000',
    verify_result TEXT NOT NULL DEFAULT '',
    process_result TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    UNIQUE(provider_code, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_provider_price_notify_log_platform_goods
    ON provider_price_notify_log(platform_account_id, supplier_goods_no);

CREATE TABLE IF NOT EXISTS open_caller (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    app_key TEXT NOT NULL,
    app_secret TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'enabled',
    allowed_ip_list TEXT NOT NULL DEFAULT '[]',
    sign_version TEXT NOT NULL DEFAULT 'v1',
    remark TEXT NOT NULL DEFAULT '',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(app_key, is_deleted)
);
CREATE INDEX IF NOT EXISTS idx_open_caller_status
    ON open_caller(status);
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
    subject_id BIGINT UNSIGNED NULL,
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
);

CREATE TABLE IF NOT EXISTS product_goods_channel_config (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    goods_id BIGINT UNSIGNED NOT NULL,
    smart_replenish_enabled TINYINT NOT NULL DEFAULT 0,
    attempt_timeout_enabled TINYINT NOT NULL DEFAULT 0,
    attempt_timeout_minutes INT NOT NULL DEFAULT 0,
    route_mode VARCHAR(32) NOT NULL DEFAULT 'fixed_order',
    sync_cost_enabled TINYINT NOT NULL DEFAULT 0,
    sync_goods_name_enabled TINYINT NOT NULL DEFAULT 0,
    allow_loss TINYINT NOT NULL DEFAULT 0,
    max_loss_amount DECIMAL(18,4) NULL,
    is_bundle TINYINT NOT NULL DEFAULT 0,
    min_channel_cost_snapshot DECIMAL(18,4) NULL,
    bound_channel_count_snapshot INT NOT NULL DEFAULT 0,
    primary_binding_id BIGINT UNSIGNED NULL,
    primary_channel_name_snapshot VARCHAR(128) NOT NULL DEFAULT '',
    channel_auto_price_status_snapshot TINYINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_goods_id (goods_id)
);

CREATE TABLE IF NOT EXISTS product_goods_channel_binding (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    goods_id BIGINT UNSIGNED NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    supplier_goods_no VARCHAR(128) NOT NULL,
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '',
    source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_direction VARCHAR(32) NOT NULL DEFAULT 'none',
    tax_adjust_rate DECIMAL(9,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    dock_status VARCHAR(16) NOT NULL DEFAULT 'enabled',
    sort INT NOT NULL DEFAULT 0,
    weight INT NOT NULL DEFAULT 0,
    start_time CHAR(5) NOT NULL DEFAULT '',
    end_time CHAR(5) NOT NULL DEFAULT '',
    validate_template_id BIGINT UNSIGNED NULL,
    is_auto_change TINYINT NOT NULL DEFAULT 0,
    add_type VARCHAR(16) NOT NULL DEFAULT 'fixed',
    default_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    lock_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    symbol_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    max_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    min_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    is_deleted TINYINT NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_goods_account_supplier_no (goods_id, platform_account_id, supplier_goods_no, is_deleted),
    KEY idx_goods_id (goods_id),
    KEY idx_platform_account_id (platform_account_id),
    KEY idx_goods_status (goods_id, dock_status, is_deleted),
    KEY idx_goods_sort (goods_id, sort, is_deleted)
);

CREATE TABLE IF NOT EXISTS trade_order (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL,
    caller_id BIGINT UNSIGNED NOT NULL,
    client_order_no VARCHAR(128) NOT NULL,
    goods_id BIGINT UNSIGNED NOT NULL,
    goods_code_snapshot VARCHAR(64) NOT NULL DEFAULT '',
    goods_name_snapshot VARCHAR(255) NOT NULL DEFAULT '',
    binding_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    platform_account_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
    route_mode_snapshot VARCHAR(32) NOT NULL DEFAULT '',
    quantity INT NOT NULL,
    success_quantity INT NOT NULL DEFAULT 0,
    failed_quantity INT NOT NULL DEFAULT 0,
    payload_json TEXT NOT NULL,
    sale_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    total_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    source_cost_price_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    cost_price_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_direction VARCHAR(32) NOT NULL DEFAULT 'none',
    tax_adjust_rate DECIMAL(9,4) NOT NULL DEFAULT 0.0000,
    tax_adjust_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    loss_order TINYINT NOT NULL DEFAULT 0,
    loss_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    channel_order_no VARCHAR(128) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'created',
    failure_reason VARCHAR(255) NOT NULL DEFAULT '',
    finished_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_order_no (order_no),
    UNIQUE KEY uk_caller_client_order_no (caller_id, client_order_no),
    KEY idx_goods_id (goods_id),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS trade_order_attempt (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    binding_id BIGINT UNSIGNED NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    provider_code VARCHAR(64) NOT NULL,
    fulfillment_no VARCHAR(64) NOT NULL,
    attempt_quantity INT NOT NULL,
    attempt_no INT NOT NULL,
    provider_request_order_no VARCHAR(128) NOT NULL,
    channel_order_no VARCHAR(128) NOT NULL DEFAULT '',
    attempt_status VARCHAR(32) NOT NULL DEFAULT 'created',
    upstream_status VARCHAR(64) NOT NULL DEFAULT '',
    binding_channel_name_snapshot VARCHAR(255) NOT NULL DEFAULT '',
    binding_supplier_goods_no_snapshot VARCHAR(128) NOT NULL DEFAULT '',
    source_cost_price_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    cost_price_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    sale_price_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    loss_amount_snapshot DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    request_url VARCHAR(512) NOT NULL DEFAULT '',
    request_method VARCHAR(16) NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL,
    request_payload TEXT NOT NULL,
    response_payload TEXT NOT NULL,
    http_status INT NOT NULL DEFAULT 0,
    duration_ms INT NOT NULL DEFAULT 0,
    error_category VARCHAR(32) NOT NULL DEFAULT '',
    error_code VARCHAR(64) NOT NULL DEFAULT '',
    error_message VARCHAR(255) NOT NULL DEFAULT '',
    query_count INT NOT NULL DEFAULT 0,
    last_query_at DATETIME NULL,
    next_query_at DATETIME NULL,
    query_deadline_at DATETIME NULL,
    callback_payload TEXT NOT NULL,
    callback_received_at DATETIME NULL,
    callback_processed_at DATETIME NULL,
    trace_id VARCHAR(128) NOT NULL DEFAULT '',
    finished_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_provider_request_order_no (provider_request_order_no),
    UNIQUE KEY uk_order_fulfillment_attempt (order_id, fulfillment_no, attempt_no),
    KEY idx_order_id (order_id),
    KEY idx_order_fulfillment (order_id, fulfillment_no),
    KEY idx_binding_id (binding_id),
    KEY idx_channel_order_no (channel_order_no),
    KEY idx_attempt_status (attempt_status),
    KEY idx_next_query_at (next_query_at)
);

CREATE TABLE IF NOT EXISTS provider_callback_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    provider_code VARCHAR(64) NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    provider_request_order_no VARCHAR(128) NOT NULL DEFAULT '',
    channel_order_no VARCHAR(128) NOT NULL DEFAULT '',
    request_headers TEXT NOT NULL,
    request_body TEXT NOT NULL,
    verify_result VARCHAR(64) NOT NULL DEFAULT '',
    process_result VARCHAR(128) NOT NULL DEFAULT '',
    ack_body TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_callback_idempotency (provider_code, idempotency_key),
    KEY idx_provider_request_order_no (provider_request_order_no),
    KEY idx_channel_order_no (channel_order_no)
);

CREATE TABLE IF NOT EXISTS provider_price_notify_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    provider_code VARCHAR(64) NOT NULL,
    platform_account_id BIGINT UNSIGNED NOT NULL,
    idempotency_key VARCHAR(128) NOT NULL,
    supplier_goods_no VARCHAR(128) NOT NULL,
    request_headers TEXT NOT NULL,
    request_body TEXT NOT NULL,
    source_cost_price_new DECIMAL(18,4) NOT NULL DEFAULT 0.0000,
    verify_result VARCHAR(64) NOT NULL DEFAULT '',
    process_result VARCHAR(128) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    UNIQUE KEY uk_price_notify_idempotency (provider_code, idempotency_key),
    KEY idx_platform_goods (platform_account_id, supplier_goods_no)
);

CREATE TABLE IF NOT EXISTS open_caller (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    app_key VARCHAR(64) NOT NULL,
    app_secret VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'enabled',
    allowed_ip_list TEXT NOT NULL,
    sign_version VARCHAR(16) NOT NULL DEFAULT 'v1',
    remark VARCHAR(255) NOT NULL DEFAULT '',
    is_deleted TINYINT NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE KEY uk_app_key (app_key, is_deleted),
    KEY idx_status (status)
);
`
