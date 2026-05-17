package services

import (
	"testing"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

func TestWorkerModule_Metadata(t *testing.T) {
	m := WorkerModule{}
	if got := m.Name(); got != "WorkerModule" {
		t.Fatalf("Name() = %q, want %q", got, "WorkerModule")
	}
	if got := m.Section(); got != "worker" {
		t.Fatalf("Section() = %q, want %q", got, "worker")
	}
}

func TestWorkerModule_LoadAndValidate(t *testing.T) {
	t.Run("unset BASE_URL uses default", func(t *testing.T) {
		t.Setenv("BASE_URL", "")
		section, err := WorkerModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.BaseURL != common.DefaultWorkerBaseURL {
			t.Fatalf("BaseURL = %q, want default %q", section.BaseURL, common.DefaultWorkerBaseURL)
		}
	})

	t.Run("custom BASE_URL", func(t *testing.T) {
		t.Setenv("BASE_URL", "https://app.example.com")
		section, err := WorkerModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.BaseURL != "https://app.example.com" {
			t.Fatalf("BaseURL = %q, want %q", section.BaseURL, "https://app.example.com")
		}
	})

	t.Run("BASE_URL whitespace is trimmed", func(t *testing.T) {
		t.Setenv("BASE_URL", "  https://app.example.com  ")
		section, err := WorkerModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.BaseURL != "https://app.example.com" {
			t.Fatalf("BaseURL = %q, want trimmed value", section.BaseURL)
		}
	})
}
