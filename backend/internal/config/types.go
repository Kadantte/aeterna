package config

import (
	"github.com/alpyxn/aeterna/backend/internal/config/common"
	"github.com/alpyxn/aeterna/backend/internal/config/services"
)

type Config struct {
	App      services.AppSection      `config:"app"`
	Database services.DatabaseSection `config:"database"`
	HTTP     services.HTTPSection     `config:"http"`
	Auth     services.AuthSection     `config:"auth"`
	Logging  services.LoggingSection  `config:"logging"`
	Worker   services.WorkerSection   `config:"worker"`
	Webhook  services.WebhookSection  `config:"webhook"`
}

type AppConfig = services.AppSection
type DatabaseConfig = services.DatabaseSection
type HTTPConfig = services.HTTPSection
type AuthConfig = services.AuthSection
type LoggingConfig = services.LoggingSection
type WorkerConfig = services.WorkerSection
type WebhookConfig = services.WebhookSection

func (c Config) IsProduction() bool {
	return c.App.Env == "production"
}

func (c Config) AllowedOriginsOrDefault() string {
	if c.HTTP.AllowedOrigins == "" {
		return common.DefaultAllowedOrigins
	}
	return c.HTTP.AllowedOrigins
}
