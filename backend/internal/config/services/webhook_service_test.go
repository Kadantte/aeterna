package services

import (
	"testing"
)

func TestWebhookModule_Metadata(t *testing.T) {
	m := WebhookModule{}
	if got := m.Name(); got != "WebhookModule" {
		t.Fatalf("Name() = %q, want %q", got, "WebhookModule")
	}
	if got := m.Section(); got != "webhook" {
		t.Fatalf("Section() = %q, want %q", got, "webhook")
	}
}

func TestWebhookModule_LoadAndValidate(t *testing.T) {
	t.Run("unset WEBHOOK_ALLOWLIST_HOSTS returns empty", func(t *testing.T) {
		t.Setenv("WEBHOOK_ALLOWLIST_HOSTS", "")
		section, err := WebhookModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowlistHosts != "" {
			t.Fatalf("AllowlistHosts = %q, want empty", section.AllowlistHosts)
		}
	})

	t.Run("custom allowlist hosts", func(t *testing.T) {
		t.Setenv("WEBHOOK_ALLOWLIST_HOSTS", "hooks.example.com,api.example.com")
		section, err := WebhookModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowlistHosts != "hooks.example.com,api.example.com" {
			t.Fatalf("AllowlistHosts = %q, want %q", section.AllowlistHosts, "hooks.example.com,api.example.com")
		}
	})

	t.Run("whitespace is trimmed", func(t *testing.T) {
		t.Setenv("WEBHOOK_ALLOWLIST_HOSTS", "  hooks.example.com  ")
		section, err := WebhookModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowlistHosts != "hooks.example.com" {
			t.Fatalf("AllowlistHosts = %q, want trimmed value", section.AllowlistHosts)
		}
	})
}
