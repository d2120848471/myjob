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
    dock_status INTEGER NOT NULL DEFAULT 1,
    sort INTEGER NOT NULL DEFAULT 0,
    order_weight TEXT NOT NULL DEFAULT '0.0000',
    order_time_start TEXT NULL,
    order_time_end TEXT NULL,
    validate_template_id INTEGER NULL,
    is_auto_change INTEGER NOT NULL DEFAULT 0,
    add_type TEXT NOT NULL DEFAULT '',
    default_price TEXT NOT NULL DEFAULT '0.0000',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_goods
    ON product_goods_channel_binding(goods_id, is_deleted, sort, id);
CREATE INDEX IF NOT EXISTS idx_product_goods_channel_binding_platform
    ON product_goods_channel_binding(platform_account_id, is_deleted, id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_product_goods_channel_binding_active
    ON product_goods_channel_binding(goods_id, platform_account_id, supplier_goods_no, is_deleted);
CREATE TABLE IF NOT EXISTS product_goods_channel_config (
    goods_id INTEGER PRIMARY KEY,
    smart_reorder_enabled INTEGER NOT NULL DEFAULT 0,
    reorder_timeout_enabled INTEGER NOT NULL DEFAULT 0,
    reorder_timeout_minutes INTEGER NOT NULL DEFAULT 0,
    order_strategy TEXT NOT NULL DEFAULT 'fixed_order',
    sync_cost_price_enabled INTEGER NOT NULL DEFAULT 0,
    sync_goods_name_enabled INTEGER NOT NULL DEFAULT 0,
    allow_loss_sale_enabled INTEGER NOT NULL DEFAULT 0,
    max_loss_amount TEXT NOT NULL DEFAULT '0.0000',
    combo_goods_enabled INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
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
    status INTEGER NOT NULL DEFAULT 1,
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
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    goods_id INTEGER NOT NULL DEFAULT 0,
    binding_id INTEGER NOT NULL DEFAULT 0,
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    callback_url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    last_action TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    subscribed_at DATETIME NULL,
    canceled_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(provider_code, platform_account_id, supplier_goods_no)
);
CREATE INDEX IF NOT EXISTS idx_supplier_product_subscription_status
    ON supplier_product_subscription(status, updated_at);
CREATE INDEX IF NOT EXISTS idx_supplier_product_subscription_goods
    ON supplier_product_subscription(goods_id, binding_id);
CREATE INDEX IF NOT EXISTS idx_supplier_product_subscription_platform
    ON supplier_product_subscription(platform_account_id, updated_at);
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    provider_code TEXT NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    binding_id INTEGER NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL DEFAULT '',
    goods_name TEXT NOT NULL DEFAULT '',
    goods_icon TEXT NOT NULL DEFAULT '',
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    old_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_source_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_cost_price TEXT NOT NULL DEFAULT '0.0000',
    new_cost_price TEXT NOT NULL DEFAULT '0.0000',
    old_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    new_effective_sell_price TEXT NOT NULL DEFAULT '0.0000',
    change_amount TEXT NOT NULL DEFAULT '0.0000',
    description TEXT NOT NULL,
    raw_payload TEXT NOT NULL,
    changed_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_price_change_log_changed
    ON product_goods_channel_price_change_log(changed_at, id);
CREATE INDEX IF NOT EXISTS idx_price_change_log_goods
    ON product_goods_channel_price_change_log(goods_id, changed_at);
CREATE INDEX IF NOT EXISTS idx_price_change_log_supplier
    ON product_goods_channel_price_change_log(provider_code, platform_account_id, supplier_goods_no, changed_at);
CREATE INDEX IF NOT EXISTS idx_price_change_log_source
    ON product_goods_channel_price_change_log(source, changed_at);
CREATE TABLE IF NOT EXISTS external_order (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_no TEXT NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL,
    goods_name TEXT NOT NULL,
    goods_type TEXT NOT NULL,
    supply_type TEXT NOT NULL,
    subject_id INTEGER NULL,
    subject_name TEXT NOT NULL DEFAULT '',
    has_tax INTEGER NOT NULL DEFAULT 0,
    account TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price TEXT NOT NULL DEFAULT '0.0000',
    order_amount TEXT NOT NULL DEFAULT '0.0000',
    cost_amount TEXT NOT NULL DEFAULT '0.0000',
    profit_amount TEXT NOT NULL DEFAULT '0.0000',
    status TEXT NOT NULL,
    current_attempt_id INTEGER NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    last_receipt TEXT NOT NULL DEFAULT '',
    next_poll_at DATETIME NULL,
    last_poll_at DATETIME NULL,
    poll_count INTEGER NOT NULL DEFAULT 0,
    last_poll_error TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_order_order_no
    ON external_order(order_no);
CREATE INDEX IF NOT EXISTS idx_external_order_status_poll
    ON external_order(status, next_poll_at, id);
CREATE INDEX IF NOT EXISTS idx_external_order_created
    ON external_order(created_at, id);
CREATE INDEX IF NOT EXISTS idx_external_order_goods
    ON external_order(goods_id, created_at);
CREATE INDEX IF NOT EXISTS idx_external_order_account
    ON external_order(account, created_at);
CREATE TABLE IF NOT EXISTS external_order_attempt (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL,
    attempt_no INTEGER NOT NULL,
    channel_binding_id INTEGER NOT NULL,
    platform_account_id INTEGER NOT NULL,
    platform_account_name TEXT NOT NULL DEFAULT '',
    platform_subject_id INTEGER NOT NULL DEFAULT 0,
    platform_subject_name TEXT NOT NULL DEFAULT '',
    provider_code TEXT NOT NULL,
    supplier_goods_no TEXT NOT NULL,
    supplier_goods_name TEXT NOT NULL DEFAULT '',
    supplier_us_order_no TEXT NOT NULL,
    supplier_order_no TEXT NOT NULL DEFAULT '',
    supplier_status TEXT NOT NULL DEFAULT '',
    refund_status TEXT NOT NULL DEFAULT '',
    request_snapshot TEXT NOT NULL,
    response_snapshot TEXT NOT NULL,
    receipt TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    submitted_at DATETIME NULL,
    last_checked_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_order_attempt_order_no
    ON external_order_attempt(order_id, attempt_no);
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_order_attempt_supplier_us
    ON external_order_attempt(provider_code, supplier_us_order_no);
CREATE INDEX IF NOT EXISTS idx_external_order_attempt_order
    ON external_order_attempt(order_id, id);
CREATE INDEX IF NOT EXISTS idx_external_order_attempt_platform
    ON external_order_attempt(platform_account_id, created_at);
CREATE TABLE IF NOT EXISTS recharge_risk_rule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account TEXT NOT NULL,
    goods_keyword TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL DEFAULT 1,
    hit_count INTEGER NOT NULL DEFAULT 0,
    created_by_id INTEGER NOT NULL DEFAULT 0,
    created_by_name TEXT NOT NULL DEFAULT '',
    updated_by_id INTEGER NOT NULL DEFAULT 0,
    updated_by_name TEXT NOT NULL DEFAULT '',
    is_deleted INTEGER NOT NULL DEFAULT 0,
    deleted_at DATETIME NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_recharge_risk_rule_active
    ON recharge_risk_rule(account, goods_keyword, is_deleted);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_match
    ON recharge_risk_rule(account, status, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_keyword
    ON recharge_risk_rule(goods_keyword, is_deleted, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_rule_status
    ON recharge_risk_rule(status, is_deleted, updated_at);
CREATE TABLE IF NOT EXISTS recharge_risk_record (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    order_no TEXT NOT NULL,
    account TEXT NOT NULL,
    goods_id INTEGER NOT NULL,
    goods_code TEXT NOT NULL,
    goods_name TEXT NOT NULL,
    matched_keyword TEXT NOT NULL,
    reason TEXT NOT NULL,
    request_token_masked TEXT NOT NULL DEFAULT '',
    intercepted_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_account
    ON recharge_risk_record(account, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_keyword
    ON recharge_risk_record(matched_keyword, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_rule
    ON recharge_risk_record(rule_id, intercepted_at, id);
CREATE INDEX IF NOT EXISTS idx_recharge_risk_record_order
    ON recharge_risk_record(order_no);
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
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    KEY idx_admin_user_group (group_id),
    KEY idx_admin_user_deleted (is_deleted)
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
    created_at DATETIME NOT NULL COMMENT '创建时间',
    KEY idx_admin_operation_log_created_at (created_at)
) COMMENT='后台操作日志表';
CREATE TABLE IF NOT EXISTS admin_login_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID',
    admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    admin_name VARCHAR(64) NOT NULL COMMENT '管理员名称',
    ip VARCHAR(45) NOT NULL COMMENT '登录IP',
    ip_region VARCHAR(128) NOT NULL COMMENT 'IP归属地',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    KEY idx_admin_login_log_created_at (created_at)
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
CREATE TABLE IF NOT EXISTS product_goods_channel_binding (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '绑定ID',
    goods_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '渠道账号ID',
    supplier_goods_no VARCHAR(128) NOT NULL COMMENT '对接商品编号',
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '对接商品名称',
    source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '原始进货价',
    cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '比较成本价',
    tax_adjust_direction VARCHAR(32) NOT NULL DEFAULT 'none' COMMENT '税额调整方向',
    tax_adjust_rate DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '税率',
    tax_adjust_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '税额调整值',
    dock_status TINYINT NOT NULL DEFAULT 1 COMMENT '对接状态',
    sort INT NOT NULL DEFAULT 0 COMMENT '排序值',
    order_weight DECIMAL(10,4) NOT NULL DEFAULT 0.0000 COMMENT '下单权重',
    order_time_start VARCHAR(5) NULL COMMENT '下单开始时段',
    order_time_end VARCHAR(5) NULL COMMENT '下单结束时段',
    validate_template_id BIGINT UNSIGNED NULL COMMENT '充值模板ID',
    is_auto_change TINYINT NOT NULL DEFAULT 0 COMMENT '是否启用自动改价',
    add_type VARCHAR(16) NOT NULL DEFAULT '' COMMENT '自动改价类型',
    default_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润值',
    is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '软删除标记',
    deleted_at DATETIME NULL COMMENT '删除时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_product_goods_channel_binding_active (goods_id, platform_account_id, supplier_goods_no, is_deleted),
    KEY idx_product_goods_channel_binding_goods (goods_id, is_deleted, sort, id),
    KEY idx_product_goods_channel_binding_platform (platform_account_id, is_deleted, id)
) COMMENT='商品渠道绑定表';
CREATE TABLE IF NOT EXISTS product_goods_channel_config (
    goods_id BIGINT UNSIGNED NOT NULL PRIMARY KEY COMMENT '商品ID',
    smart_reorder_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '智能补单开关',
    reorder_timeout_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '补单超时开关',
    reorder_timeout_minutes INT NOT NULL DEFAULT 0 COMMENT '补单超时分钟数',
    order_strategy VARCHAR(32) NOT NULL DEFAULT 'fixed_order' COMMENT '下单策略',
    sync_cost_price_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '同步进价开关',
    sync_goods_name_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '同步商品名称开关',
    allow_loss_sale_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '亏本销售开关',
    max_loss_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '允许亏本金额',
    combo_goods_enabled TINYINT NOT NULL DEFAULT 0 COMMENT '组合商品开关',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='商品渠道库存配置表';
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
CREATE TABLE IF NOT EXISTS supplier_product_subscription (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订阅记录ID',
    provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
    platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
    platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
    goods_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '本地商品ID',
    binding_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '渠道绑定ID',
    supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
    callback_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '订阅回调地址',
    status VARCHAR(32) NOT NULL COMMENT '订阅状态',
    last_action VARCHAR(32) NOT NULL DEFAULT '' COMMENT '最近动作',
    last_error VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
    request_snapshot TEXT NOT NULL COMMENT '最近请求快照',
    response_snapshot TEXT NOT NULL COMMENT '最近响应快照',
    subscribed_at DATETIME NULL COMMENT '最近订阅成功时间',
    canceled_at DATETIME NULL COMMENT '最近取消成功时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_supplier_product_subscription_active (provider_code, platform_account_id, supplier_goods_no),
    KEY idx_supplier_product_subscription_status (status, updated_at),
    KEY idx_supplier_product_subscription_goods (goods_id, binding_id),
    KEY idx_supplier_product_subscription_platform (platform_account_id, updated_at)
) COMMENT='供应商商品订阅记录表';
CREATE TABLE IF NOT EXISTS product_goods_channel_price_change_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '改价记录ID',
    source VARCHAR(32) NOT NULL COMMENT '来源',
    provider_code VARCHAR(32) NOT NULL COMMENT '供应商适配器编码',
    platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
    platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
    binding_id BIGINT UNSIGNED NOT NULL COMMENT '渠道绑定ID',
    goods_id BIGINT UNSIGNED NOT NULL COMMENT '本地商品ID',
    goods_code VARCHAR(32) NOT NULL DEFAULT '' COMMENT '本地商品编码快照',
    goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '本地商品名称快照',
    goods_icon VARCHAR(500) NOT NULL DEFAULT '' COMMENT '商品图标快照',
    supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
    old_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前原始进货价',
    new_source_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后原始进货价',
    old_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前比较成本价',
    new_cost_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后比较成本价',
    old_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动前利润后价格',
    new_effective_sell_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '变动后利润后价格',
    change_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润后价格变化值',
    description TEXT NOT NULL COMMENT '变动描述',
    raw_payload TEXT NOT NULL COMMENT '原始载荷',
    changed_at DATETIME NOT NULL COMMENT '变动时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    KEY idx_price_change_log_changed (changed_at, id),
    KEY idx_price_change_log_goods (goods_id, changed_at),
    KEY idx_price_change_log_supplier (provider_code, platform_account_id, supplier_goods_no, changed_at),
    KEY idx_price_change_log_source (source, changed_at)
) COMMENT='商品渠道自动改价记录表';
CREATE TABLE IF NOT EXISTS external_order (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '订单ID',
    order_no VARCHAR(40) NOT NULL COMMENT '订单号',
    goods_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
    goods_code VARCHAR(32) NOT NULL COMMENT '商品编码快照',
    goods_name VARCHAR(255) NOT NULL COMMENT '商品名称快照',
    goods_type VARCHAR(32) NOT NULL COMMENT '商品类型快照',
    supply_type VARCHAR(32) NOT NULL COMMENT '供货方式快照',
    subject_id BIGINT UNSIGNED NULL COMMENT '商品主体ID快照',
    subject_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '商品主体名称快照',
    has_tax TINYINT NOT NULL DEFAULT 0 COMMENT '商品含税快照',
    account VARCHAR(255) NOT NULL COMMENT '充值账号',
    quantity INT NOT NULL COMMENT '购买数量',
    unit_price DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '下单单价',
    order_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '订单金额',
    cost_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '成本金额',
    profit_amount DECIMAL(18,4) NOT NULL DEFAULT 0.0000 COMMENT '利润金额',
    status VARCHAR(32) NOT NULL COMMENT '订单状态',
    current_attempt_id BIGINT UNSIGNED NULL COMMENT '当前尝试ID',
    attempt_count INT NOT NULL DEFAULT 0 COMMENT '尝试次数',
    last_receipt VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近回执',
    next_poll_at DATETIME NULL COMMENT '下次查单时间',
    last_poll_at DATETIME NULL COMMENT '最近查单时间',
    poll_count INT NOT NULL DEFAULT 0 COMMENT '查单次数',
    last_poll_error VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近查单异常',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_external_order_order_no (order_no),
    KEY idx_external_order_status_poll (status, next_poll_at, id),
    KEY idx_external_order_created (created_at, id),
    KEY idx_external_order_goods (goods_id, created_at),
    KEY idx_external_order_account (account, created_at)
) COMMENT='外部订单主表';
CREATE TABLE IF NOT EXISTS external_order_attempt (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '尝试ID',
    order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    order_no VARCHAR(40) NOT NULL COMMENT '订单号',
    attempt_no INT NOT NULL COMMENT '尝试序号',
    channel_binding_id BIGINT UNSIGNED NOT NULL COMMENT '商品渠道绑定ID',
    platform_account_id BIGINT UNSIGNED NOT NULL COMMENT '平台账号ID',
    platform_account_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '平台账号名称快照',
    platform_subject_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '平台账号主体ID快照',
    platform_subject_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '平台账号主体名称快照',
    provider_code VARCHAR(32) NOT NULL COMMENT '适配器编码',
    supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
    supplier_goods_name VARCHAR(255) NOT NULL DEFAULT '' COMMENT '上游商品名称快照',
    supplier_us_order_no VARCHAR(64) NOT NULL COMMENT '上游商家单号',
    supplier_order_no VARCHAR(128) NOT NULL DEFAULT '' COMMENT '上游订单号',
    supplier_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游原始状态',
    refund_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游退款状态',
    request_snapshot TEXT NOT NULL COMMENT '请求快照',
    response_snapshot TEXT NOT NULL COMMENT '响应快照',
    receipt VARCHAR(512) NOT NULL DEFAULT '' COMMENT '上游回执',
    status VARCHAR(32) NOT NULL COMMENT '尝试状态',
    submitted_at DATETIME NULL COMMENT '提交时间',
    last_checked_at DATETIME NULL COMMENT '最近查单时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_external_order_attempt_order_no (order_id, attempt_no),
    UNIQUE KEY uk_external_order_attempt_supplier_us (provider_code, supplier_us_order_no),
    KEY idx_external_order_attempt_order (order_id, id),
    KEY idx_external_order_attempt_platform (platform_account_id, created_at)
) COMMENT='外部订单渠道尝试表';
CREATE TABLE IF NOT EXISTS recharge_risk_rule (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '风控规则ID',
    account VARCHAR(255) NOT NULL COMMENT '充值账号',
    goods_keyword VARCHAR(255) NOT NULL COMMENT '商品名关键词',
    reason VARCHAR(512) NOT NULL DEFAULT '' COMMENT '风控原因',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用，0停用',
    hit_count INT NOT NULL DEFAULT 0 COMMENT '累计拦截次数',
    created_by_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人ID快照',
    created_by_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人名称快照',
    updated_by_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '更新人ID快照',
    updated_by_name VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人名称快照',
    is_deleted TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除',
    deleted_at DATETIME NULL COMMENT '删除时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间',
    UNIQUE KEY uk_recharge_risk_rule_active (account, goods_keyword, is_deleted),
    KEY idx_recharge_risk_rule_match (account, status, is_deleted, id),
    KEY idx_recharge_risk_rule_keyword (goods_keyword, is_deleted, id),
    KEY idx_recharge_risk_rule_status (status, is_deleted, updated_at)
) COMMENT='充值账号风控规则表';
CREATE TABLE IF NOT EXISTS recharge_risk_record (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '拦截记录ID',
    rule_id BIGINT UNSIGNED NOT NULL COMMENT '命中规则ID',
    order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
    order_no VARCHAR(40) NOT NULL COMMENT '订单号',
    account VARCHAR(255) NOT NULL COMMENT '充值账号',
    goods_id BIGINT UNSIGNED NOT NULL COMMENT '商品ID快照',
    goods_code VARCHAR(32) NOT NULL COMMENT '商品编码快照',
    goods_name VARCHAR(255) NOT NULL COMMENT '商品名称快照',
    matched_keyword VARCHAR(255) NOT NULL COMMENT '命中关键词快照',
    reason VARCHAR(512) NOT NULL COMMENT '风控原因快照',
    request_token_masked VARCHAR(64) NOT NULL DEFAULT '' COMMENT '开放下单token脱敏快照',
    intercepted_at DATETIME NOT NULL COMMENT '拦截时间',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    KEY idx_recharge_risk_record_account (account, intercepted_at, id),
    KEY idx_recharge_risk_record_keyword (matched_keyword, intercepted_at, id),
    KEY idx_recharge_risk_record_rule (rule_id, intercepted_at, id),
    KEY idx_recharge_risk_record_order (order_no)
) COMMENT='充值账号风控拦截记录表';
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
    config_key VARCHAR(64) NOT NULL UNIQUE COMMENT '配置键',
    config_value TEXT NOT NULL COMMENT '配置值',
    description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置描述',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    updated_at DATETIME NOT NULL COMMENT '更新时间'
) COMMENT='系统配置表'
`
