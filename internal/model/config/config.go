package config

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/util/gconv"
)

type Config struct {
	AppEnv    string           `json:"app_env" yaml:"app_env"`
	Server    ServerConfig     `json:"server" yaml:"server"`
	Database  DatabaseConfig   `json:"database" yaml:"database"`
	Redis     RedisConfig      `json:"redis" yaml:"redis"`
	Auth      AuthConfig       `json:"auth" yaml:"auth"`
	Bootstrap BootstrapConfig  `json:"bootstrap" yaml:"bootstrap"`
	SMS       RuntimeSMSConfig `json:"sms" yaml:"sms"`
	Audit     AuditConfig      `json:"audit" yaml:"audit"`
	Upload    UploadConfig     `json:"upload" yaml:"upload"`
	OpenOrder OpenOrderConfig  `json:"open_order" yaml:"open_order"`
}

type ServerConfig struct {
	Address string `json:"address" yaml:"address"`
}

type DatabaseConfig struct {
	Driver string `json:"driver" yaml:"driver"`
	DSN    string `json:"dsn" yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `json:"addr" yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

type AuthConfig struct {
	JWTSecret         string `json:"jwt_secret" yaml:"jwt_secret"`
	AccessTokenTTLMin int    `json:"access_token_ttl_minutes" yaml:"access_token_ttl_minutes"`
	TempLoginTTLMin   int    `json:"temp_login_ttl_minutes" yaml:"temp_login_ttl_minutes"`
}

type BootstrapConfig struct {
	SuperAdminUsername string `json:"super_admin_username" yaml:"super_admin_username"`
	SuperAdminPhone    string `json:"super_admin_phone" yaml:"super_admin_phone"`
	SuperAdminPassword string `json:"super_admin_password" yaml:"super_admin_password"`
}

type RuntimeSMSConfig struct {
	Provider string `json:"provider" yaml:"provider"`
}

type AuditConfig struct {
	Async      bool `json:"async" yaml:"async"`
	BufferSize int  `json:"buffer_size" yaml:"buffer_size"`
}

type UploadConfig struct {
	LocalDir       string `json:"local_dir" yaml:"local_dir"`
	PublicPrefix   string `json:"public_prefix" yaml:"public_prefix"`
	MaxImageSizeMB int    `json:"max_image_size_mb" yaml:"max_image_size_mb"`
}

type OpenOrderConfig struct {
	Token                     string `json:"token" yaml:"token"`
	WorkerEnabled             bool   `json:"worker_enabled" yaml:"worker_enabled"`
	PollIntervalSeconds       int    `json:"poll_interval_seconds" yaml:"poll_interval_seconds"`
	SubmitScanIntervalSeconds int    `json:"submit_scan_interval_seconds" yaml:"submit_scan_interval_seconds"`
}

func Default() Config {
	return Config{
		AppEnv: "dev",
		Server: ServerConfig{Address: ":8080"},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local",
		},
		Redis:     RedisConfig{Addr: "127.0.0.1:6380", DB: 0},
		Auth:      AuthConfig{JWTSecret: "please-change-this-secret", AccessTokenTTLMin: 10080, TempLoginTTLMin: 5},
		Bootstrap: BootstrapConfig{SuperAdminUsername: "admin", SuperAdminPhone: "15881767197", SuperAdminPassword: "abc123"},
		SMS:       RuntimeSMSConfig{Provider: "aliyun"},
		Audit:     AuditConfig{Async: true, BufferSize: 128},
		Upload:    UploadConfig{LocalDir: "storage/uploads", PublicPrefix: "/uploads", MaxImageSizeMB: 2},
		OpenOrder: OpenOrderConfig{
			Token:                     "test-open-order-token",
			WorkerEnabled:             false,
			PollIntervalSeconds:       30,
			SubmitScanIntervalSeconds: 5,
		},
	}
}

func LoadFromGoFrame(ctx context.Context, cfg *gcfg.Config) (Config, error) {
	result := Default()
	data, err := cfg.Data(ctx)
	if err != nil {
		return result, err
	}
	if len(data) > 0 {
		if err = gconv.Scan(data, &result); err != nil {
			return result, err
		}
	}
	normalize(&result)
	return result, nil
}

func Normalize(cfg *Config) {
	normalize(cfg)
}

func normalize(cfg *Config) {
	cfg.AppEnv = strings.TrimSpace(cfg.AppEnv)
	cfg.Server.Address = expand(cfg.Server.Address)
	cfg.Database.Driver = strings.TrimSpace(expand(cfg.Database.Driver))
	cfg.Database.DSN = expand(cfg.Database.DSN)
	cfg.Redis.Addr = expand(cfg.Redis.Addr)
	cfg.Redis.Password = expand(cfg.Redis.Password)
	cfg.Auth.JWTSecret = expand(cfg.Auth.JWTSecret)
	cfg.Bootstrap.SuperAdminUsername = strings.TrimSpace(expand(cfg.Bootstrap.SuperAdminUsername))
	cfg.Bootstrap.SuperAdminPhone = strings.TrimSpace(expand(cfg.Bootstrap.SuperAdminPhone))
	cfg.Bootstrap.SuperAdminPassword = strings.TrimSpace(expand(cfg.Bootstrap.SuperAdminPassword))
	cfg.SMS.Provider = strings.TrimSpace(expand(cfg.SMS.Provider))
	cfg.Upload.LocalDir = expand(cfg.Upload.LocalDir)
	cfg.Upload.PublicPrefix = strings.TrimSpace(expand(cfg.Upload.PublicPrefix))
	cfg.OpenOrder.Token = strings.TrimSpace(expand(cfg.OpenOrder.Token))
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "mysql"
	}
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8080"
	}
	if cfg.Auth.AccessTokenTTLMin <= 0 {
		cfg.Auth.AccessTokenTTLMin = 10080
	}
	if cfg.Auth.TempLoginTTLMin <= 0 {
		cfg.Auth.TempLoginTTLMin = 5
	}
	if cfg.Audit.BufferSize <= 0 {
		cfg.Audit.BufferSize = 128
	}
	if cfg.Upload.LocalDir == "" {
		cfg.Upload.LocalDir = "storage/uploads"
	}
	if cfg.Upload.PublicPrefix == "" {
		cfg.Upload.PublicPrefix = "/uploads"
	}
	if !strings.HasPrefix(cfg.Upload.PublicPrefix, "/") {
		cfg.Upload.PublicPrefix = "/" + cfg.Upload.PublicPrefix
	}
	cfg.Upload.PublicPrefix = strings.TrimRight(cfg.Upload.PublicPrefix, "/")
	if cfg.Upload.PublicPrefix == "" {
		cfg.Upload.PublicPrefix = "/uploads"
	}
	if cfg.Upload.MaxImageSizeMB <= 0 {
		cfg.Upload.MaxImageSizeMB = 2
	}
	if cfg.OpenOrder.PollIntervalSeconds <= 0 {
		cfg.OpenOrder.PollIntervalSeconds = 30
	}
	if cfg.OpenOrder.SubmitScanIntervalSeconds <= 0 {
		cfg.OpenOrder.SubmitScanIntervalSeconds = 5
	}
	if cfg.Bootstrap.SuperAdminUsername == "" {
		cfg.Bootstrap.SuperAdminUsername = "admin"
	}
}

var envDefaultRegexp = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*):-([^}]*)}`)

func expand(value string) string {
	if value == "" {
		return value
	}
	// 本地 YAML 使用 ${VAR:-default} 形式，Go 标准库只支持 ${VAR}，这里先补齐默认值语义。
	value = envDefaultRegexp.ReplaceAllStringFunc(value, func(match string) string {
		parts := envDefaultRegexp.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		if envValue := os.Getenv(parts[1]); envValue != "" {
			return envValue
		}
		return parts[2]
	})
	return os.ExpandEnv(value)
}
