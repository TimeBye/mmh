package core

import (
	"path/filepath"
	"testing"
	"time"
)

func TestConfigLoadMergeDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "default.yaml")
	cfg := &Config{
		Basic: BasicServerConfig{User: "admin"},
		Servers: Servers{
			{Name: "s1", Address: "127.0.0.1"},
		},
	}
	if err := cfg.WriteTo(path); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loaded := &Config{}
	if err := loaded.LoadFrom(path); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Servers) != 1 {
		t.Fatalf("got %d servers, want 1", len(loaded.Servers))
	}

	s := loaded.Servers[0]
	if s.User != "admin" {
		t.Errorf("user = %q, want admin", s.User)
	}
	if s.Port != 22 {
		t.Errorf("port = %d, want 22", s.Port)
	}
	if s.ServerAliveInterval != 10*time.Second {
		t.Errorf("server_alive_interval = %s, want 10s", s.ServerAliveInterval)
	}
	if s.ConfigPath != path {
		t.Errorf("config path = %q, want %q", s.ConfigPath, path)
	}
}
