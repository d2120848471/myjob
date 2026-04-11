package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

func defaultConfig() Config {
	return Config{
		AppEnv: "dev",
		Server: ServerConfig{Address: ":8080"},
		Database: DatabaseConfig{
			Driver: "mysql",
			DSN:    "root:root123456@tcp(127.0.0.1:3306)/admin?charset=utf8mb4&parseTime=true&loc=Local",
		},
		Redis:     RedisConfig{Addr: "127.0.0.1:6379", DB: 0},
		Auth:      AuthConfig{JWTSecret: "dev-secret-change-me", AccessTokenTTLMin: 10080, TempLoginTTLMin: 5},
		Bootstrap: BootstrapConfig{SuperAdminUsername: "admin"},
		SMS:       RuntimeSMSConfig{Provider: "mock"},
		Audit:     AuditConfig{Async: true, BufferSize: 128},
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := defaultConfig()
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
