package app

import (
	"context"
	"database/sql"
	"strings"

	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"golang.org/x/crypto/bcrypt"
)

// execStatements 将 schema 按分号拆分为多条 SQL 并依次执行。
//
// 用于执行内置建表语句，避免引入额外的 migration 依赖。
func execStatements(ctx context.Context, exec func(string, ...any) (sql.Result, error), schema string) error {
	parts := strings.Split(schema, ";")
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		if _, err := exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// bootstrap 执行启动引导：建表、补列、写入默认种子数据与默认配置。
//
// 该流程设计为幂等，可重复执行（用于首次启动或版本升级时补齐缺失数据）。
func (c *Core) bootstrap(ctx context.Context) error {
	if c.driver == "sqlite" {
		// sqlite 与 mysql 的建表语法不同，按驱动选择对应 schema。
		if err := execStatements(ctx, func(sql string, args ...any) (sql.Result, error) { return c.DB().Exec(ctx, sql, args...) }, sqliteSchema); err != nil {
			return err
		}
	} else {
		if err := execStatements(ctx, func(sql string, args ...any) (sql.Result, error) { return c.DB().Exec(ctx, sql, args...) }, mysqlSchema); err != nil {
			return err
		}
	}
	if err := c.ensureMenuSchema(ctx); err != nil {
		return err
	}
	if err := c.ensureProductGoodsSchema(ctx); err != nil {
		return err
	}
	if err := c.ensureDefaultGroup(ctx); err != nil {
		return err
	}
	if err := c.ensureMenus(ctx); err != nil {
		return err
	}
	if err := c.ensureDefaultGroupAuth(ctx); err != nil {
		return err
	}
	if err := c.ensureSupplierPlatformTypes(ctx); err != nil {
		return err
	}
	if err := c.ensureSMSConfig(ctx); err != nil {
		return err
	}
	if err := c.ensureSuperAdmin(ctx); err != nil {
		return err
	}
	return nil
}

// ensureMenuSchema 为菜单表补齐历史版本可能缺失的列。
func (c *Core) ensureMenuSchema(ctx context.Context) error {
	definitions := map[string]string{"menu_level": "INTEGER NOT NULL DEFAULT 1", "status": "INTEGER NOT NULL DEFAULT 1", "super_only": "INTEGER NOT NULL DEFAULT 0"}
	if c.driver == "sqlite" {
		rows := make([]struct {
			Name string `db:"name"`
		}, 0)
		if err := c.DB().GetCore().GetScan(ctx, &rows, `PRAGMA table_info(admin_menu)`); err != nil {
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
			if _, err := c.DB().Exec(ctx, `ALTER TABLE admin_menu ADD COLUMN `+column+` `+definition); err != nil {
				return err
			}
		}
		return nil
	}
	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	if err := c.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM admin_menu`); err != nil {
		return err
	}
	existing := map[string]struct{}{}
	for _, row := range rows {
		existing[row.Field] = struct{}{}
	}
	mysqlDefinitions := map[string]string{"menu_level": "TINYINT NOT NULL DEFAULT 1", "status": "TINYINT NOT NULL DEFAULT 1", "super_only": "TINYINT NOT NULL DEFAULT 0"}
	for column, definition := range mysqlDefinitions {
		if _, ok := existing[column]; ok {
			continue
		}
		if _, err := c.DB().Exec(ctx, `ALTER TABLE admin_menu ADD COLUMN `+column+` `+definition); err != nil {
			return err
		}
	}
	return nil
}

// ensureProductGoodsSchema 为商品表补齐 subject_id 字段（用于主体配置关联）。
func (c *Core) ensureProductGoodsSchema(ctx context.Context) error {
	if c.driver == "sqlite" {
		rows := make([]struct {
			Name string `db:"name"`
		}, 0)
		if err := c.DB().GetCore().GetScan(ctx, &rows, `PRAGMA table_info(product_goods)`); err != nil {
			return err
		}
		for _, row := range rows {
			if row.Name == "subject_id" {
				return nil
			}
		}
		_, err := c.DB().Exec(ctx, `ALTER TABLE product_goods ADD COLUMN subject_id INTEGER NULL`)
		return err
	}

	rows := make([]struct {
		Field string `db:"Field"`
	}, 0)
	if err := c.DB().GetCore().GetScan(ctx, &rows, `SHOW COLUMNS FROM product_goods`); err != nil {
		return err
	}
	for _, row := range rows {
		if row.Field == "subject_id" {
			return nil
		}
	}
	_, err := c.DB().Exec(ctx, `ALTER TABLE product_goods ADD COLUMN subject_id BIGINT UNSIGNED NULL`)
	return err
}

// ensureDefaultGroup 确保默认用户组（id=1）存在。
func (c *Core) ensureDefaultGroup(ctx context.Context) error {
	count, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group WHERE id = 1`)
	if err != nil {
		return err
	}
	if count.Int() > 0 {
		return nil
	}
	_, err = c.DB().Exec(ctx, `INSERT INTO admin_group (id, name, description, status, created_at, updated_at) VALUES (1, '默认组', '默认权限组', 1, ?, ?)`, c.now(), c.now())
	return err
}

// defaultMenus 返回系统内置菜单/权限点种子数据。
func defaultMenus() []menuSeed {
	return []menuSeed{
		{ID: 1, ParentID: 0, Name: "员工管理", Code: "admin.list", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 1},
		{ID: 2, ParentID: 0, Name: "用户组与授权", Code: "admin.department", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 2},
		{ID: 3, ParentID: 0, Name: "操作日志", Code: "admin.action", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 3},
		{ID: 4, ParentID: 0, Name: "登录日志", Code: "admin.loginlog", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 4},
		{ID: 5, ParentID: 0, Name: "主体配置", Code: "subject.manage", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 5},
		{ID: 6, ParentID: 0, Name: "短信配置", Code: "config.sms", MenuLevel: 1, Status: 1, SuperOnly: 1, Sort: 6},
		{ID: 7, ParentID: 0, Name: "品牌管理", Code: "product.brand", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 7},
		{ID: 8, ParentID: 0, Name: "行业管理", Code: "product.industry", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 8},
		{ID: 9, ParentID: 0, Name: "系统参数配置", Code: "config.system", MenuLevel: 1, Status: 1, SuperOnly: 1, Sort: 9},
		{ID: 10, ParentID: 0, Name: "商品模板管理", Code: "product.template", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 10},
		{ID: 11, ParentID: 0, Name: "商品购买数量限制策略", Code: "product.purchase_limit", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 11},
		{ID: 12, ParentID: 0, Name: "第三方对接", Code: "supplier.index", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 12},
		{ID: 13, ParentID: 0, Name: "商品管理", Code: "product.goods", MenuLevel: 1, Status: 1, SuperOnly: 0, Sort: 13},
	}
}

// ensureMenus 写入或更新内置菜单数据（用于新增权限点或修正排序/名称）。
func (c *Core) ensureMenus(ctx context.Context) error {
	for _, item := range defaultMenus() {
		exists, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_menu WHERE id = ?`, item.ID)
		if err != nil {
			return err
		}
		if exists.Int() == 0 {
			if _, err = c.DB().Exec(ctx, `INSERT INTO admin_menu (id, parent_id, name, code, menu_type, menu_level, status, super_only, sort, created_at, updated_at) VALUES (?, ?, ?, ?, 'permission', ?, ?, ?, ?, ?, ?)`, item.ID, item.ParentID, item.Name, item.Code, item.MenuLevel, item.Status, item.SuperOnly, item.Sort, c.now(), c.now()); err != nil {
				return err
			}
			continue
		}
		if _, err = c.DB().Exec(ctx, `UPDATE admin_menu SET parent_id = ?, name = ?, code = ?, menu_type = 'permission', menu_level = ?, status = ?, super_only = ?, sort = ?, updated_at = ? WHERE id = ?`, item.ParentID, item.Name, item.Code, item.MenuLevel, item.Status, item.SuperOnly, item.Sort, c.now(), item.ID); err != nil {
			return err
		}
	}
	return nil
}

// ensureDefaultGroupAuth 为默认用户组补齐初始授权菜单。
func (c *Core) ensureDefaultGroupAuth(ctx context.Context) error {
	for _, menuID := range []int64{1, 2, 3, 4, 5, 7, 8, 10, 11, 12, 13} {
		exists, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_group_menu WHERE group_id = 1 AND menu_id = ?`, menuID)
		if err != nil {
			return err
		}
		if exists.Int() > 0 {
			continue
		}
		if _, err = c.DB().Exec(ctx, `INSERT INTO admin_group_menu (group_id, menu_id, created_at) VALUES (1, ?, ?)`, menuID, c.now()); err != nil {
			return err
		}
	}
	return nil
}

// ensureSupplierPlatformTypes 初始化第三方平台类型字典（如不同供应商平台分类）。
func (c *Core) ensureSupplierPlatformTypes(ctx context.Context) error {
	for _, item := range supplierprovider.DefaultTypes() {
		exists, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM supplier_platform_type WHERE id = ?`, item.ID)
		if err != nil {
			return err
		}
		if exists.Int() == 0 {
			if _, err = c.DB().Exec(ctx, `INSERT INTO supplier_platform_type (id, type_name, default_provider_code, status, sort, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, item.ID, item.TypeName, item.ProviderCode, item.Status, item.Sort, c.now(), c.now()); err != nil {
				return err
			}
			continue
		}
		if _, err = c.DB().Exec(ctx, `UPDATE supplier_platform_type SET type_name = ?, default_provider_code = ?, status = ?, sort = ?, updated_at = ? WHERE id = ?`, item.TypeName, item.ProviderCode, item.Status, item.Sort, c.now(), item.ID); err != nil {
			return err
		}
	}
	return nil
}

// ensureSMSConfig 初始化短信配置默认项（写入 system_config 表）。
func (c *Core) ensureSMSConfig(ctx context.Context) error {
	defaults := map[string]string{"sms_access_key": "", "sms_access_key_secret": "", "sms_sign_name": "玖权益", "sms_template_code": "SMS_000001", "sms_expire_minutes": "30", "sms_interval_minutes": "1"}
	for key, value := range defaults {
		exists, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key)
		if err != nil {
			return err
		}
		if exists.Int() > 0 {
			continue
		}
		if _, err = c.DB().Exec(ctx, `INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, key, value, key, c.now(), c.now()); err != nil {
			return err
		}
	}
	return nil
}

// ensureSuperAdmin 确保超级管理员账号存在且始终同步本地固定引导凭证。
// 这样即使旧库里已有 admin，也会在启动期回写手机号和密码，避免本地调试凭证漂移。
func (c *Core) ensureSuperAdmin(ctx context.Context) error {
	username := c.cfg.Bootstrap.SuperAdminUsername
	if username == "" {
		username = "admin"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.cfg.Bootstrap.SuperAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	count, err := c.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_user WHERE username = ?`, username)
	if err != nil {
		return err
	}
	if count.Int() > 0 {
		_, err = c.DB().Exec(ctx, `
UPDATE admin_user
SET password_hash = ?, real_name = ?, phone = ?, group_id = 0, status = 1, is_deleted = 0, is_business = 1, deleted_at = NULL, token_version = token_version + 1, updated_at = ?
WHERE username = ?
`, string(hash), "系统管理员", c.cfg.Bootstrap.SuperAdminPhone, c.now(), username)
		return err
	}
	_, err = c.DB().Exec(ctx, `
INSERT INTO admin_user (
    username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted,
    last_login_ip, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, 0, 1, 0, 1, 0, ?, 0, ?, ?)
`, username, string(hash), "系统管理员", c.cfg.Bootstrap.SuperAdminPhone, "127.0.0.1", c.now(), c.now())
	return err
}

// bcryptGenerate 生成 bcrypt 密码哈希（用于创建用户/测试用户等）。
func bcryptGenerate(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
