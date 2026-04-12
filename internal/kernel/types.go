package kernel

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"myjob/internal/library/audit"
	"myjob/internal/library/region"
	"myjob/internal/library/sms"
	"myjob/internal/model/config"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/redis/go-redis/v9"
)

var (
	usernameRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{5,9}$`)
	passwordRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{5,9}$`)
	phoneRegexp    = regexp.MustCompile(`^1\d{10}$`)
	smsCodeRegexp  = regexp.MustCompile(`^\d{6}$`)
)

type Config = config.Config

type Core struct {
	cfg            config.Config
	driver         string
	db             gdb.DB
	redis          *redis.Client
	now            func() time.Time
	sender         sms.Sender
	mock           *sms.MockSender
	regionResolver region.RegionResolver
	auditWriter    *audit.Writer
	tempDBFile     string
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

func (c *Core) Config() config.Config {
	return c.cfg
}

func (c *Core) DB() gdb.DB {
	return c.db
}

func (c *Core) Redis() *redis.Client {
	return c.redis
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
	result, err := c.db.Exec(ctx, `
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
