package services

import (
	"strings"
	"testing"
)

func TestHTTPModule_Metadata(t *testing.T) {
	m := HTTPModule{}
	if got := m.Name(); got != "HTTPModule" {
		t.Fatalf("Name() = %q, want %q", got, "HTTPModule")
	}
	if got := m.Section(); got != "http" {
		t.Fatalf("Section() = %q, want %q", got, "http")
	}
}

func TestHTTPModule_LoadAndValidate(t *testing.T) {
	t.Run("development mode with no env vars", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("ALLOWED_ORIGINS", "")
		t.Setenv("PROXY_MODE", "")
		section, err := HTTPModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowedOriginsIsSet {
			t.Fatal("AllowedOriginsIsSet should be false when ALLOWED_ORIGINS is empty")
		}
		if section.AllowedOrigins != "" {
			t.Fatalf("AllowedOrigins = %q, want empty", section.AllowedOrigins)
		}
	})

	t.Run("custom ALLOWED_ORIGINS", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("ALLOWED_ORIGINS", "https://example.com")
		t.Setenv("PROXY_MODE", "")
		section, err := HTTPModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowedOrigins != "https://example.com" {
			t.Fatalf("AllowedOrigins = %q, want %q", section.AllowedOrigins, "https://example.com")
		}
		if !section.AllowedOriginsIsSet {
			t.Fatal("AllowedOriginsIsSet should be true")
		}
	})

	t.Run("production requires ALLOWED_ORIGINS", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("ALLOWED_ORIGINS", "")
		t.Setenv("PROXY_MODE", "")
		_, err := HTTPModule{}.LoadAndValidate()
		if err == nil {
			t.Fatal("expected error when ALLOWED_ORIGINS is unset in production")
		}
		if !strings.Contains(err.Error(), "ALLOWED_ORIGINS") {
			t.Fatalf("error should mention ALLOWED_ORIGINS, got: %v", err)
		}
	})

	t.Run("production with wildcard ALLOWED_ORIGINS fails without simple mode", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("ALLOWED_ORIGINS", "*")
		t.Setenv("PROXY_MODE", "")
		_, err := HTTPModule{}.LoadAndValidate()
		if err == nil {
			t.Fatal("expected error for wildcard ALLOWED_ORIGINS in production without simple mode")
		}
		if !strings.Contains(err.Error(), "*") {
			t.Fatalf("error should mention wildcard, got: %v", err)
		}
	})

	t.Run("production with wildcard ALLOWED_ORIGINS and simple proxy mode succeeds", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("ALLOWED_ORIGINS", "*")
		t.Setenv("PROXY_MODE", "simple")
		section, err := HTTPModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.AllowedOrigins != "*" {
			t.Fatalf("AllowedOrigins = %q, want %q", section.AllowedOrigins, "*")
		}
		if section.ProxyMode != "simple" {
			t.Fatalf("ProxyMode = %q, want %q", section.ProxyMode, "simple")
		}
	})

	t.Run("production with specific origins succeeds", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
		t.Setenv("PROXY_MODE", "")
		section, err := HTTPModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !section.AllowedOriginsIsSet {
			t.Fatal("AllowedOriginsIsSet should be true")
		}
	})

	t.Run("PROXY_MODE is captured", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("ALLOWED_ORIGINS", "")
		t.Setenv("PROXY_MODE", "simple")
		section, err := HTTPModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.ProxyMode != "simple" {
			t.Fatalf("ProxyMode = %q, want %q", section.ProxyMode, "simple")
		}
	})
}
