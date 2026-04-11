package main

import (
	"log"
	"net/http"
	"os"

	"admin/internal/app"
)

func main() {
	configPath := os.Getenv("ADMIN_CONFIG")
	if configPath == "" {
		configPath = "manifest/config/config.local.yaml"
		if _, err := os.Stat(configPath); err != nil {
			configPath = "manifest/config/config.example.yaml"
		}
	}
	cfg, err := app.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	application, err := app.NewApplicationFromConfig(cfg)
	if err != nil {
		log.Fatalf("create application failed: %v", err)
	}
	defer application.Close()

	log.Printf("admin backend listening on %s", cfg.Server.Address)
	if err = http.ListenAndServe(cfg.Server.Address, application.Handler()); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
