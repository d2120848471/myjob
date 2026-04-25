package app

import "context"

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
