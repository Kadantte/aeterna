package services

import (
	"testing"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/config"
)

func TestSessionTTL(t *testing.T) {
	tests := []struct {
		name  string
		hours int
		want  time.Duration
	}{
		{"twelve hours", 12, 12 * time.Hour},
		{"six hours", 6, 6 * time.Hour},
		{"one hour", 1, time.Hour},
		{"large value", 720, 720 * time.Hour},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewAuthService(config.Config{Auth: config.AuthConfig{SessionTTLHours: tc.hours}})
			if got := svc.sessionTTL(); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
