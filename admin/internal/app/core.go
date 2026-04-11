package app

import (
	"admin/internal/model/entity"
	"admin/utility/ipx"
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	statusEnabled  = 1
	statusDisabled = 0
)

var (
	usernameRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{5,9}$`)
	passwordRegexp = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{5,9}$`)
	phoneRegexp    = regexp.MustCompile(`^1\d{10}$`)
	smsCodeRegexp  = regexp.MustCompile(`^\d{6}$`)
)

type Config struct {
	AppEnv    string           `yaml:"app_env"`
	Server    ServerConfig     `yaml:"server"`
	Database  DatabaseConfig   `yaml:"database"`
	Redis     RedisConfig      `yaml:"redis"`
	Auth      AuthConfig       `yaml:"auth"`
	Bootstrap BootstrapConfig  `yaml:"bootstrap"`
	SMS       RuntimeSMSConfig `yaml:"sms"`
	Audit     AuditConfig      `yaml:"audit"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AuthConfig struct {
	JWTSecret         string `yaml:"jwt_secret"`
	AccessTokenTTLMin int    `yaml:"access_token_ttl_minutes"`
	TempLoginTTLMin   int    `yaml:"temp_login_ttl_minutes"`
}

type BootstrapConfig struct {
	SuperAdminUsername string `yaml:"super_admin_username"`
	SuperAdminPhone    string `yaml:"super_admin_phone"`
	SuperAdminPassword string `yaml:"super_admin_password"`
}

type RuntimeSMSConfig struct {
	Provider string `yaml:"provider"`
}

type AuditConfig struct {
	Async      bool `yaml:"async"`
	BufferSize int  `yaml:"buffer_size"`
}

type Application struct {
	cfg            Config
	driver         string
	db             *sqlx.DB
	redis          *redis.Client
	mux            *http.ServeMux
	now            func() time.Time
	sender         SMSSender
	syncAudit      bool
	auditCh        chan operationEvent
	auditWG        sync.WaitGroup
	auditStop      chan struct{}
	mock           *MockSMSSender
	regionResolver ipx.RegionResolver
}

type AdminUser = entity.AdminUser
type userListItem = entity.UserListItem
type AdminGroup = entity.AdminGroup
type groupListItem = entity.GroupListItem
type AdminMenu = entity.AdminMenu
type AdminSubject = entity.AdminSubject

type SMSConfig struct {
	AccessKey       string `json:"access_key"`
	AccessKeySecret string `json:"access_key_secret"`
	SignName        string `json:"sign_name"`
	TemplateCode    string `json:"template_code"`
	ExpireMinutes   int    `json:"expire_minutes"`
	IntervalMinutes int    `json:"interval_minutes"`
}

type SMSConfigSaveRequest struct {
	AccessKey           string `json:"access_key"`
	AccessKeySecret     string `json:"access_key_secret"`
	SignName            string `json:"sign_name"`
	TemplateCode        string `json:"template_code"`
	ExpireMinutes       int    `json:"expire_minutes"`
	IntervalMinutes     int    `json:"interval_minutes"`
	KeepAccessKey       bool   `json:"keep_access_key"`
	KeepAccessKeySecret bool   `json:"keep_access_key_secret"`
}

type SMSConfigGetResponse struct {
	AccessKeyMasked           string `json:"access_key_masked"`
	AccessKeySecretMasked     string `json:"access_key_secret_masked"`
	AccessKeyConfigured       bool   `json:"access_key_configured"`
	AccessKeySecretConfigured bool   `json:"access_key_secret_configured"`
	SignName                  string `json:"sign_name"`
	TemplateCode              string `json:"template_code"`
	ExpireMinutes             int    `json:"expire_minutes"`
	IntervalMinutes           int    `json:"interval_minutes"`
	UpdatedAt                 string `json:"updated_at,omitempty"`
}

type smsConfigState struct {
	Version                   int       `json:"version"`
	Config                    SMSConfig `json:"config"`
	AccessKeyConfigured       bool      `json:"access_key_configured"`
	AccessKeySecretConfigured bool      `json:"access_key_secret_configured"`
	UpdatedAt                 time.Time `json:"updated_at,omitempty"`
}

type operationLog = entity.OperationLog
type loginLog = entity.LoginLog

type responseEnvelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type principal struct {
	UserID       int64
	GroupID      int64
	TokenVersion int
	JTI          string
}

type jwtClaims struct {
	UserID       int64 `json:"user_id"`
	GroupID      int64 `json:"group_id"`
	TokenVersion int   `json:"token_version"`
	jwt.RegisteredClaims
}

type sessionPayload struct {
	UserID       int64     `json:"user_id"`
	GroupID      int64     `json:"group_id"`
	TokenVersion int       `json:"token_version"`
	JTI          string    `json:"jti"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type tempLoginPayload struct {
	UserID   int64  `json:"user_id"`
	IP       string `json:"ip"`
	Attempts int    `json:"attempts"`
}

type smsCodePayload struct {
	LoginToken string `json:"login_token"`
	Code       string `json:"code"`
}

type operationEvent struct {
	AdminID     int64
	AdminName   string
	Description string
	IP          string
	IPRegion    string
}

type SMSSender interface {
	SendLoginCode(ctx context.Context, phone, code string, cfg SMSConfig) error
}

type MockSMSSender struct {
	mu    sync.RWMutex
	codes map[string]string
}

type apiError struct {
	HTTPStatus int
	Code       int
	Message    string
}

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

func (e *apiError) Error() string {
	return e.Message
}

func newMockSMSSender() *MockSMSSender {
	return &MockSMSSender{codes: make(map[string]string)}
}

func (m *MockSMSSender) SendLoginCode(_ context.Context, phone, code string, _ SMSConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[phone] = code
	return nil
}

func (m *MockSMSSender) LastCode(phone string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	code, ok := m.codes[phone]
	if !ok {
		return "", errors.New("sms code not found")
	}
	return code, nil
}
