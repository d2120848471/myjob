package app

import (
	"context"
	"database/sql"
)

// ensureExternalOrderAttemptSchema 为历史订单尝试表补齐渠道主体快照列。
func (c *Core) ensureExternalOrderAttemptSchema(ctx context.Context) error {
	definitions := map[string]string{
		"platform_subject_id":   "INTEGER NOT NULL DEFAULT 0",
		"platform_subject_name": "TEXT NOT NULL DEFAULT ''",
	}
	if c.driver == "sqlite" {
		rows := make([]struct {
			Name string `db:"name"`
		}, 0)
		if err := c.DB().GetCore().GetScan(ctx, &rows, `PRAGMA table_info(external_order_attempt)`); err != nil {
			return err
		}
		existing := map[string]struct{}{}
		for _, row := range rows {
			existing[row.Name] = struct{}{}
		}
		for column, definition := range definitions {
			if _, ok := existing[column]; ok {
				continue
			}
			if _, err := c.DB().Exec(ctx, `ALTER TABLE external_order_attempt ADD COLUMN `+column+` `+definition); err != nil {
				return err
			}
		}
		return nil
	}

	mysqlDefinitions := map[string]string{
		"platform_subject_id":   "BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '平台账号主体ID快照'",
		"platform_subject_name": "VARCHAR(64) NOT NULL DEFAULT '' COMMENT '平台账号主体名称快照'",
	}
	// 订单尝试表可能持续写入，启动补列只作为小表/本地升级兜底；短锁等待避免长时间卡住启动。
	if _, err := c.DB().Exec(ctx, `SET SESSION lock_wait_timeout = 5`); err != nil {
		return err
	}
	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	if err := c.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM external_order_attempt`); err != nil {
		return err
	}
	existing := map[string]struct{}{}
	for _, row := range rows {
		existing[row.Field] = struct{}{}
	}
	for column, definition := range mysqlDefinitions {
		if _, ok := existing[column]; ok {
			continue
		}
		if _, err := c.DB().Exec(ctx, `ALTER TABLE external_order_attempt ADD COLUMN `+column+` `+definition); err != nil {
			return err
		}
	}
	return nil
}

const sqliteExternalOrderAttemptSegmentSchema = `
CREATE TABLE IF NOT EXISTS external_order_attempt_segment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    attempt_id INTEGER NOT NULL,
    segment_no INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    provider_code TEXT NOT NULL,
    supplier_goods_no TEXT NOT NULL,
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
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_order_attempt_segment_no
    ON external_order_attempt_segment(attempt_id, segment_no);
CREATE UNIQUE INDEX IF NOT EXISTS uk_external_order_segment_supplier_us
    ON external_order_attempt_segment(provider_code, supplier_us_order_no);
CREATE INDEX IF NOT EXISTS idx_external_order_segment_attempt
    ON external_order_attempt_segment(attempt_id, id);
CREATE INDEX IF NOT EXISTS idx_external_order_segment_order
    ON external_order_attempt_segment(order_id, id);
`

const mysqlExternalOrderAttemptSegmentSchema = `
CREATE TABLE IF NOT EXISTS external_order_attempt_segment (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '尝试子单ID',
  order_id BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  attempt_id BIGINT UNSIGNED NOT NULL COMMENT '尝试ID',
  segment_no INT NOT NULL COMMENT '子单序号',
  quantity INT NOT NULL COMMENT '子单数量',
  provider_code VARCHAR(32) NOT NULL COMMENT '适配器编码',
  supplier_goods_no VARCHAR(128) NOT NULL COMMENT '上游商品编号',
  supplier_us_order_no VARCHAR(80) NOT NULL COMMENT '上游商家单号',
  supplier_order_no VARCHAR(128) NOT NULL DEFAULT '' COMMENT '上游订单号',
  supplier_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游原始状态',
  refund_status VARCHAR(32) NOT NULL DEFAULT '' COMMENT '上游退款状态',
  request_snapshot TEXT NOT NULL COMMENT '请求快照',
  response_snapshot TEXT NOT NULL COMMENT '响应快照',
  receipt VARCHAR(512) NOT NULL DEFAULT '' COMMENT '上游回执',
  status VARCHAR(32) NOT NULL COMMENT '子单状态',
  submitted_at DATETIME NULL COMMENT '提交时间',
  last_checked_at DATETIME NULL COMMENT '最近查单时间',
  created_at DATETIME NOT NULL COMMENT '创建时间',
  updated_at DATETIME NOT NULL COMMENT '更新时间',
  UNIQUE KEY uk_external_order_attempt_segment_no (attempt_id, segment_no),
  UNIQUE KEY uk_external_order_segment_supplier_us (provider_code, supplier_us_order_no),
  KEY idx_external_order_segment_attempt (attempt_id, id),
  KEY idx_external_order_segment_order (order_id, id)
) COMMENT='外部订单渠道尝试子单表';
`

// ensureExternalOrderAttemptSegmentSchema 为支持拆单的平台创建真实上游请求粒度的子单表。
func (c *Core) ensureExternalOrderAttemptSegmentSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		return execStatements(ctx, func(sql string, args ...any) (sql.Result, error) {
			return c.DB().Exec(ctx, sql, args...)
		}, sqliteExternalOrderAttemptSegmentSchema)
	}
	// 子单表是订单提交路径新增表；启动兜底建表设置短锁等待，避免生产大库元数据锁长期阻塞。
	if _, err := c.DB().Exec(ctx, `SET SESSION lock_wait_timeout = 5`); err != nil {
		return err
	}
	return execStatements(ctx, func(sql string, args ...any) (sql.Result, error) {
		return c.DB().Exec(ctx, sql, args...)
	}, mysqlExternalOrderAttemptSegmentSchema)
}
