package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	TableAdminUser         = "admin_user"
	TableAdminGroup        = "admin_group"
	TableAdminMenu         = "admin_menu"
	TableAdminGroupMenu    = "admin_group_menu"
	TableAdminOperationLog = "admin_operation_log"
	TableAdminLoginLog     = "admin_login_log"
	TableAdminSubject      = "admin_subject"
	TableProductGoods      = "product_goods"
	TableSystemConfig      = "system_config"
)

func AdminUserModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminUser).Ctx(ctx).Safe()
}
func AdminGroupModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminGroup).Ctx(ctx).Safe()
}
func AdminMenuModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminMenu).Ctx(ctx).Safe()
}
func AdminGroupMenuModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminGroupMenu).Ctx(ctx).Safe()
}
func AdminOperationLogModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminOperationLog).Ctx(ctx).Safe()
}
func AdminLoginLogModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminLoginLog).Ctx(ctx).Safe()
}
func AdminSubjectModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableAdminSubject).Ctx(ctx).Safe()
}
func ProductGoodsModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableProductGoods).Ctx(ctx).Safe()
}
func SystemConfigModel(db gdb.DB, ctx context.Context) *gdb.Model {
	return db.Model(TableSystemConfig).Ctx(ctx).Safe()
}
