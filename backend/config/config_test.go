package config

import (
	"testing"
	"time"
)

func TestConfigShutdownTimeoutDefault(t *testing.T) {
	cfg := Load()
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("缺省关停超时应为 10s, got %v", cfg.ShutdownTimeout)
	}
}
