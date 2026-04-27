package app

import (
	"context"
	"database/sql"
)

const sqliteRechargeRiskSchema = `
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
`

const mysqlRechargeRiskSchema = `
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
`

// ensureRechargeRiskSchema 确保充值风控规则和拦截流水表存在。
func (c *Core) ensureRechargeRiskSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		return execStatements(ctx, func(sqlText string, args ...any) (sql.Result, error) {
			return c.DB().Exec(ctx, sqlText, args...)
		}, sqliteRechargeRiskSchema)
	}
	return execStatements(ctx, func(sqlText string, args ...any) (sql.Result, error) {
		return c.DB().Exec(ctx, sqlText, args...)
	}, mysqlRechargeRiskSchema)
}
