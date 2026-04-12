package config

import (
	"os"

	"gopkg.in/yaml.v3"
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

func Default() Config {
	return Config{
		AppEnv: "dev",
		Server: ServerConfig{Address: ":8080"},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "root:root123456@tcp(127.0.0.1:3307)/admin?charset=utf8mb4&parseTime=true&loc=Local",
		},
		Redis:     RedisConfig{Addr: "127.0.0.1:6380", DB: 0},
		Auth:      AuthConfig{JWTSecret: "please-change-this-secret", AccessTokenTTLMin: 10080, TempLoginTTLMin: 5},
		Bootstrap: BootstrapConfig{SuperAdminUsername: "admin"},
		SMS:       RuntimeSMSConfig{Provider: "mock"},
		Audit:     AuditConfig{Async: true, BufferSize: 128},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	expanded := os.ExpandEnv(string(data))
	if err = yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return cfg, err
	}
	// 配置文件允许只覆盖少量字段，缺失项继续走默认值兜底。
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
	return cfg, nil
}
