package app

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"myjob/internal/library/audit"
	"myjob/internal/library/region"
	"myjob/internal/library/sms"
	modelconfig "myjob/internal/model/config"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"

	"github.com/alicebob/miniredis/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

var (
	usernameRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{5,9}$`)
	passwordRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{5,9}$`)
	phoneRegexp    = regexp.MustCompile(`^1\d{10}$`)
	smsCodeRegexp  = regexp.MustCompile(`^\d{6}$`)
)

// Config 是后台运行时配置结构（来自 internal/model/config）。
type Config = modelconfig.Config

// Core 是后台运行时核心对象，负责持有配置与基础设施依赖（DB/Redis/SMS/审计等）。
type Core struct {
	cfg            modelconfig.Config
	cfgName        string
	cfgInstance    *gcfg.Config
	dbGroup        string
	redisGroup     string
	driver         string
	now            func() time.Time
	sender         sms.Sender
	mock           *sms.MockSender
	regionResolver region.RegionResolver
	auditWriter    *audit.Writer
	tempDBFile     string
	tempUploadDir  string
	miniRedis      *miniredis.Miniredis
}

// AdminUser 是后台管理用户实体的类型别名。
type AdminUser = entity.AdminUser

// UserListItem 是员工列表页条目结构的类型别名。
type UserListItem = entity.UserListItem

// AdminGroup 是后台用户组实体的类型别名。
type AdminGroup = entity.AdminGroup

// GroupListItem 是用户组列表条目结构的类型别名。
type GroupListItem = entity.GroupListItem

// AdminMenu 是后台菜单/权限点实体的类型别名。
type AdminMenu = entity.AdminMenu

// AdminSubject 是主体配置实体的类型别名。
type AdminSubject = entity.AdminSubject

// ProductBrand 是商品品牌实体的类型别名。
type ProductBrand = entity.ProductBrand

// BrandListItem 是品牌列表条目结构的类型别名。
type BrandListItem = entity.BrandListItem

// ProductIndustry 是商品行业实体的类型别名。
type ProductIndustry = entity.ProductIndustry

// IndustryListItem 是行业列表条目结构的类型别名。
type IndustryListItem = entity.IndustryListItem

// IndustryBrandRelationItem 是行业-品牌关联条目的类型别名。
type IndustryBrandRelationItem = entity.IndustryBrandRelationItem

// BrandSelectorItem 是行业绑定品牌下拉选择条目的类型别名。
type BrandSelectorItem = entity.BrandSelectorItem

// OperationLog 是操作日志实体的类型别名。
type OperationLog = entity.OperationLog

// LoginLog 是登录日志实体的类型别名。
type LoginLog = entity.LoginLog

type smsConfigState = modelruntime.SMSConfigState

type menuSeed struct {
	ID        int64
	ParentID  int64
	Name      string
	Code      string
	MenuLevel int
	Status    int
	SuperOnly int
	Sort      int
}

type sqlExecutor interface {
	Exec(sql string, args ...any) (sql.Result, error)
	GetValue(sql string, args ...any) (gdb.Value, error)
	GetArray(sql string, args ...any) (gdb.Array, error)
	GetScan(pointer any, sql string, args ...any) error
	GetAll(sql string, args ...any) (gdb.Result, error)
}

// Config 返回当前 Core 的配置快照（只读值）。
func (c *Core) Config() modelconfig.Config {
	return c.cfg
}

// Cfg 返回 GoFrame 配置实例，用于读取配置文件内容。
func (c *Core) Cfg() *gcfg.Config {
	return c.cfgInstance
}

// DB 返回 Core 绑定的数据库实例（使用独立 group 避免与全局配置冲突）。
func (c *Core) DB() gdb.DB {
	return g.DB(c.dbGroup)
}

// Redis 返回 Core 绑定的 Redis 实例（使用独立 group 避免与全局配置冲突）。
func (c *Core) Redis() *gredis.Redis {
	return gredis.Instance(c.redisGroup)
}

// SetSMSSender 替换短信发送实现（用于测试或切换 provider）。
func (c *Core) SetSMSSender(sender sms.Sender) {
	c.sender = sender
	if mock, ok := sender.(*sms.MockSender); ok {
		c.mock = mock
		return
	}
	c.mock = nil
}

// LastMockSMSCode 返回 mock 短信发送器最近一次发送的验证码。
//
// 当 sender 不是 mock 时返回 sql.ErrNoRows。
func (c *Core) LastMockSMSCode(phone string) (string, error) {
	if c.mock == nil {
		return "", sql.ErrNoRows
	}
	return c.mock.LastCode(phone)
}

// CreateTestUser 在当前数据库中创建一个测试用户（用于单测/集成测试）。
func (c *Core) CreateTestUser(ctx context.Context, username, password, phone string) (int64, error) {
	hash, err := bcryptGenerate(password)
	if err != nil {
		return 0, err
	}
	result, err := c.DB().Exec(ctx, `
INSERT INTO admin_user (
    username, password_hash, real_name, phone, group_id, status, balance_notify, is_business, is_deleted, token_version, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, ?)
`, username, hash, username, phone, 1, 1, c.now(), c.now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Now 返回 Core 内部使用的当前时间（便于测试注入）。
func (c *Core) Now() time.Time {
	return c.now()
}

// Sender 返回当前 Core 使用的短信发送器实现。
func (c *Core) Sender() sms.Sender {
	return c.sender
}

// UsernameRegexp 返回用户名校验正则（字母开头，长度 6-10，仅字母数字）。
func UsernameRegexp() *regexp.Regexp { return usernameRegexp }

// PasswordRegexp 返回密码校验正则（字母开头，长度 6-10，允许下划线）。
func PasswordRegexp() *regexp.Regexp { return passwordRegexp }

// PhoneRegexp 返回手机号校验正则（1 开头 11 位）。
func PhoneRegexp() *regexp.Regexp { return phoneRegexp }

// SMSCodeRegexp 返回短信验证码校验正则（6 位数字）。
func SMSCodeRegexp() *regexp.Regexp { return smsCodeRegexp }
