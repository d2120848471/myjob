package app

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func execStatements(ctx context.Context, db *sqlx.DB, schema string) error {
	parts := strings.Split(schema, ";")
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) ensureMenuSchema(ctx context.Context) error {
	definitions := map[string]string{
		"menu_level": "INTEGER NOT NULL DEFAULT 1",
		"status":     "INTEGER NOT NULL DEFAULT 1",
		"super_only": "INTEGER NOT NULL DEFAULT 0",
	}
	if a.driver == "sqlite" {
		rows := make([]struct {
			CID        int            `db:"cid"`
			Name       string         `db:"name"`
			Type       string         `db:"type"`
			NotNull    int            `db:"notnull"`
			Default    sql.NullString `db:"dflt_value"`
			PrimaryKey int            `db:"pk"`
		}, 0)
		if err := a.db.SelectContext(ctx, &rows, `PRAGMA table_info(admin_menu)`); err != nil {
			return err
		}
		existing := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			existing[row.Name] = struct{}{}
		}
		for column, definition := range definitions {
			if _, ok := existing[column]; ok {
				continue
			}
			if _, err := a.db.ExecContext(ctx, `ALTER TABLE admin_menu ADD COLUMN `+column+` `+definition); err != nil {
				return err
			}
		}
		return nil
	}
	rows := make([]struct {
		Field   string         `db:"Field"`
		Type    string         `db:"Type"`
		Null    string         `db:"Null"`
		Key     string         `db:"Key"`
		Default sql.NullString `db:"Default"`
		Extra   string         `db:"Extra"`
	}, 0)
	if err := a.db.SelectContext(ctx, &rows, `SHOW COLUMNS FROM admin_menu`); err != nil {
		return err
	}
	existing := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		existing[row.Field] = struct{}{}
	}
	mysqlDefinitions := map[string]string{
		"menu_level": "TINYINT NOT NULL DEFAULT 1",
		"status":     "TINYINT NOT NULL DEFAULT 1",
		"super_only": "TINYINT NOT NULL DEFAULT 0",
	}
	for column, definition := range mysqlDefinitions {
		if _, ok := existing[column]; ok {
			continue
		}
		if _, err := a.db.ExecContext(ctx, `ALTER TABLE admin_menu ADD COLUMN `+column+` `+definition); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) ensureDefaultGroup(ctx context.Context) error {
	var exists int
	if err := a.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM admin_group WHERE id = 1`); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	_, err := a.db.ExecContext(ctx, `
INSERT INTO admin_group (id, name, description, status, created_at, updated_at)
VALUES (1, '默认组', '默认权限组', 1, ?, ?)
`, a.now(), a.now())
	return err
}

func defaultMenus() []menuSeed {
	return []menuSeed{
		{ID: 1, ParentID: 0, Name: "员工管理", Code: "admin.list", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 1},
		{ID: 2, ParentID: 0, Name: "用户组与授权", Code: "admin.department", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 2},
		{ID: 3, ParentID: 0, Name: "操作日志", Code: "admin.action", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 3},
		{ID: 4, ParentID: 0, Name: "登录日志", Code: "admin.loginlog", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 4},
		{ID: 5, ParentID: 0, Name: "主体配置", Code: "subject.manage", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 5},
		{ID: 6, ParentID: 0, Name: "短信配置", Code: "config.sms", MenuLevel: 1, Status: 1, SuperOnly: 1, Sort: 6},
	}
}

func (a *Application) ensureMenus(ctx context.Context) error {
	for _, item := range defaultMenus() {
		var exists int
		if err := a.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM admin_menu WHERE id = ?`, item.ID); err != nil {
			return err
		}
		if exists == 0 {
			if _, err := a.db.ExecContext(ctx, `
INSERT INTO admin_menu (id, parent_id, name, code, menu_type, menu_level, status, super_only, sort, created_at, updated_at)
VALUES (?, ?, ?, ?, 'permission', ?, ?, ?, ?, ?, ?)
`, item.ID, item.ParentID, item.Name, item.Code, item.MenuLevel, item.Status, item.SuperOnly, item.Sort, a.now(), a.now()); err != nil {
				return err
			}
			continue
		}
		if _, err := a.db.ExecContext(ctx, `
UPDATE admin_menu
SET parent_id = ?, name = ?, code = ?, menu_type = 'permission', menu_level = ?, status = ?, super_only = ?, sort = ?, updated_at = ?
WHERE id = ?
`, item.ParentID, item.Name, item.Code, item.MenuLevel, item.Status, item.SuperOnly, item.Sort, a.now(), item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) ensureDefaultGroupAuth(ctx context.Context) error {
	var exists int
	if err := a.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM admin_group_menu WHERE group_id = 1`); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	for _, menuID := range []int64{1, 2, 3, 4, 5} {
		if _, err := a.db.ExecContext(ctx, `INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (1, ?, ?)`, menuID, a.now()); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) ensureSMSConfig(ctx context.Context) error {
	defaults := map[string]string{
		"sms_access_key":        "",
		"sms_access_key_secret": "",
		"sms_sign_name":         "玖权益",
		"sms_template_code":     "SMS_000001",
		"sms_expire_minutes":    "30",
		"sms_interval_minutes":  "1",
	}
	for key, value := range defaults {
		var exists int
		if err := a.db.GetContext(ctx, &exists, `SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		if _, err := a.db.ExecContext(ctx, `
INSERT INTO system_config (config_key, config_value, description, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
`, key, value, key, a.now(), a.now()); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) ensureSuperAdmin(ctx context.Context) error {
	username := a.cfg.Bootstrap.SuperAdminUsername
	if username == "" {
		username = "admin"
	}
	var count int
	if err := a.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM admin_user WHERE username = ?`, username); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(a.cfg.Bootstrap.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.db.ExecContext(ctx, `
INSERT INTO admin_user (
    username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
    last_login_ip, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, 0, 1, 0, 1, 0, ?, 0, ?, ?)
`, username, string(hash), "系统管理员", a.cfg.Bootstrap.SuperAdminPhone, "127.0.0.1", a.now(), a.now())
	return err
}

func (a *Application) startAuditWorker() {
	if a.syncAudit {
		return
	}
	a.auditWG.Add(1)
	go func() {
		defer a.auditWG.Done()
		for {
			select {
			case <-a.auditStop:
				return
			case evt := <-a.auditCh:
				_ = a.insertOperationLog(context.Background(), evt)
			}
		}
	}()
}
