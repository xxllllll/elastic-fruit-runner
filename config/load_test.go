package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithArgs_DefaultAPIAddr(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("idle_timeout: 15m\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadWithArgs([]string{"--config", path})
	if err != nil {
		t.Fatalf("LoadWithArgs() error: %v", err)
	}
	if cfg.APIAddr != DefaultAPIAddr {
		t.Fatalf("APIAddr = %q, want %q", cfg.APIAddr, DefaultAPIAddr)
	}
}
