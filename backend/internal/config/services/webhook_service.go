package services

import (
	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

type WebhookModule struct{}

func (WebhookModule) Name() string { return "WebhookModule" }
func (WebhookModule) Section() string {
	return "webhook"
}

func init() {
	common.Register(WebhookModule{})
}

type WebhookSection struct {
	AllowlistHosts string
}

func (WebhookModule) LoadAndValidate() (WebhookSection, error) {
	return WebhookSection{
		AllowlistHosts: common.GetenvTrim("WEBHOOK_ALLOWLIST_HOSTS"),
	}, nil
}
