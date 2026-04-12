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

type Config = modelconfig.Config

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
	miniRedis      *miniredis.Miniredis
}

type AdminUser = entity.AdminUser

type UserListItem = entity.UserListItem

type AdminGroup = entity.AdminGroup

type GroupListItem = entity.GroupListItem

type AdminMenu = entity.AdminMenu

type AdminSubject = entity.AdminSubject

type OperationLog = entity.OperationLog

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

func (c *Core) Config() modelconfig.Config {
	return c.cfg
}

func (c *Core) Cfg() *gcfg.Config {
	return c.cfgInstance
}

func (c *Core) DB() gdb.DB {
	return g.DB(c.dbGroup)
}

func (c *Core) Redis() *gredis.Redis {
	return gredis.Instance(c.redisGroup)
}

func (c *Core) SetSMSSender(sender sms.Sender) {
	c.sender = sender
	if mock, ok := sender.(*sms.MockSender); ok {
		c.mock = mock
		return
	}
	c.mock = nil
}

func (c *Core) LastMockSMSCode(phone string) (string, error) {
	if c.mock == nil {
		return "", sql.ErrNoRows
	}
	return c.mock.LastCode(phone)
}

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

func (c *Core) Now() time.Time {
	return c.now()
}

func (c *Core) Sender() sms.Sender {
	return c.sender
}

func UsernameRegexp() *regexp.Regexp { return usernameRegexp }
func PasswordRegexp() *regexp.Regexp { return passwordRegexp }
func PhoneRegexp() *regexp.Regexp    { return phoneRegexp }
func SMSCodeRegexp() *regexp.Regexp  { return smsCodeRegexp }
