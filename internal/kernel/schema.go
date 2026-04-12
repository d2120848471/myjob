package kernel

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
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(64) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
)
`
