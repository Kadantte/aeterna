package services

import (
	"testing"
)

func TestAppModule_Metadata(t *testing.T) {
	m := AppModule{}
	if got := m.Name(); got != "AppModule" {
		t.Fatalf("Name() = %q, want %q", got, "AppModule")
	}
	if got := m.Section(); got != "app" {
		t.Fatalf("Section() = %q, want %q", got, "app")
	}
}

func TestAppModule_LoadAndValidate(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		wantEnv string
	}{
		{"unset ENV returns empty", "", ""},
		{"development mode", "development", "development"},
		{"production mode", "production", "production"},
		{"custom value", "staging", "staging"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENV", tc.envVal)
			section, err := AppModule{}.LoadAndValidate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if section.Env != tc.wantEnv {
				t.Fatalf("Env = %q, want %q", section.Env, tc.wantEnv)
			}
		})
	}
}
