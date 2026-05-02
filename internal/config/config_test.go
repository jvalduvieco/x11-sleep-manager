package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultUsesRuntimeDirWhenSet(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/tmp/runtime-test")

	cfg := Default()

	if got, want := cfg.Socket.Path, "/tmp/runtime-test/x11_sleep_manager.sock"; got != want {
		t.Fatalf("socket path mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Reconcile.Interval.Duration, 15*time.Second; got != want {
		t.Fatalf("interval mismatch: got %s want %s", got, want)
	}
	if !cfg.X11.DisableScreenSaver || !cfg.X11.DisableDPMS || !cfg.X11.RestorePreviousState {
		t.Fatal("expected X11 defaults to be enabled")
	}
	if got, want := strings.Join(cfg.Processes.Pause, ","), "xss-lock"; got != want {
		t.Fatalf("unexpected pause processes: got %q want %q", got, want)
	}
}

func TestLoadEmptyPathReturnsDefaults(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/tmp/runtime-test")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := cfg.Match.Who, []string{"OpenCode"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("who mismatch: got %v want %v", got, want)
	}
}

func TestLoadMergesWithDefaults(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/tmp/runtime-test")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
	  "socket": {"path": "/tmp/custom.sock"},
	  "reconcile": {"interval": "30s"},
	  "logging": {"level": "debug", "format": "text"}
	}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := cfg.Socket.Path, "/tmp/custom.sock"; got != want {
		t.Fatalf("socket path mismatch: got %q want %q", got, want)
	}
	if got, want := cfg.Reconcile.Interval.Duration, 30*time.Second; got != want {
		t.Fatalf("interval mismatch: got %s want %s", got, want)
	}
	if got, want := cfg.Match.Who, []string{"OpenCode"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("default match.who lost: got %v want %v", got, want)
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"reconcile": {"interval": "0s"}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "reconcile.interval") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyPatchReplacesSupportedSections(t *testing.T) {
	cfg := Default()
	next, err := cfg.ApplyPatch(ConfigPatch{
		Match: &MatchConfig{UID: "self", Who: []string{"Editor"}, WhatAny: []string{"idle"}},
		X11:   &X11Config{DisableScreenSaver: false, DisableDPMS: true, RestorePreviousState: true},
	})
	if err != nil {
		t.Fatalf("ApplyPatch returned error: %v", err)
	}
	if got, want := next.Match.Who[0], "Editor"; got != want {
		t.Fatalf("unexpected match patch result: %+v", next.Match)
	}
	if next.X11.DisableScreenSaver {
		t.Fatal("expected x11 patch to be applied")
	}
	if got, want := next.Processes.Pause[0], "xss-lock"; got != want {
		t.Fatalf("unexpected untouched processes config: %+v", next.Processes)
	}
}
