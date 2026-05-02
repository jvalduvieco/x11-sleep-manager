package control

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

type fakeRuntimeController struct {
	reconcileErr   error
	enableErr      error
	disableErr     error
	reconcileCalls int
	enableCalls    int
	disableCalls   int
}

func (f *fakeRuntimeController) Reconcile(context.Context) error {
	f.reconcileCalls++
	return f.reconcileErr
}

func (f *fakeRuntimeController) Enable(context.Context) error {
	f.enableCalls++
	return f.enableErr
}

func (f *fakeRuntimeController) Disable(context.Context) error {
	f.disableCalls++
	return f.disableErr
}

func TestStatusEndpointServesSnapshot(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	store.RegisterSession(state.Session{Display: ":0", XAuthority: "/tmp/.Xauthority"})

	server := NewServerWithSessionRegistration(socketPath, store, store)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}

	resp, err := client.Get("http://unix/v1/status")
	if err != nil {
		t.Fatalf("get status: %v", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("status code mismatch: got %d want %d", got, want)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var snapshot state.Snapshot
	if err := json.Unmarshal(body, &snapshot); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if snapshot.State != state.ModeIdle {
		t.Fatalf("unexpected state: %s", snapshot.State)
	}
	if !snapshot.SessionReady {
		t.Fatal("expected session ready to be true")
	}
	if snapshot.Session == nil || snapshot.Session.Display != ":0" {
		t.Fatalf("unexpected session payload: %+v", snapshot.Session)
	}
	if snapshot.Version != "test" {
		t.Fatalf("unexpected version: %s", snapshot.Version)
	}
}

func TestRuntimeControlEndpointsInvokeController(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	runtime := &fakeRuntimeController{}
	server := NewServerWithRuntimeControl(socketPath, store, store, runtime)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}}

	for _, path := range []string{"/v1/reconcile", "/v1/enable", "/v1/disable"} {
		resp, err := client.Post("http://unix"+path, "application/json", nil)
		if err != nil {
			t.Fatalf("post %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("unexpected status for %s: %s", path, resp.Status)
		}
	}
	if runtime.reconcileCalls != 1 || runtime.enableCalls != 1 || runtime.disableCalls != 1 {
		t.Fatalf("unexpected runtime calls: %+v", runtime)
	}
}

func TestRuntimeControlEndpointPropagatesError(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	runtime := &fakeRuntimeController{reconcileErr: errors.New("boom")}
	server := NewServerWithRuntimeControl(socketPath, store, store, runtime)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}}
	resp, err := client.Post("http://unix/v1/reconcile", "application/json", nil)
	if err != nil {
		t.Fatalf("post reconcile: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
}

func TestConfigEndpointServesEffectiveConfig(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	server := NewServerWithSessionRegistration(socketPath, store, store)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}}
	resp, err := client.Get("http://unix/v1/config")
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	var cfg config.Config
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if got, want := cfg.Match.Who[0], "OpenCode"; got != want {
		t.Fatalf("unexpected config payload: %+v", cfg)
	}
}

func TestSessionRegistrationEndpointStoresSession(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	server := NewServerWithSessionRegistration(socketPath, store, store)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}

	body := `{"display":":1","xauthority":"/tmp/auth","xdg_session_type":"x11"}`
	resp, err := client.Post("http://unix/v1/session", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post session: %v", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusNoContent; got != want {
		t.Fatalf("status code mismatch: got %d want %d", got, want)
	}

	snapshot := store.Snapshot()
	if !snapshot.SessionReady {
		t.Fatal("expected session ready after registration")
	}
	if snapshot.Session == nil || snapshot.Session.Display != ":1" {
		t.Fatalf("unexpected registered session: %+v", snapshot.Session)
	}
}

func TestSessionRegistrationRejectsInvalidPayload(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	server := NewServerWithSessionRegistration(socketPath, store, store)
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}

	body := `{"display":"","xauthority":"/tmp/auth"}`
	resp, err := client.Post("http://unix/v1/session", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post session: %v", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("status code mismatch: got %d want %d", got, want)
	}
	if store.Snapshot().SessionReady {
		t.Fatal("session should not be registered on invalid payload")
	}
}

func TestShutdownRemovesSocket(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	server := NewServer(socketPath, state.NewStore(config.Default(), "test"))
	if err := server.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown server: %v", err)
	}

	if _, err := net.Dial("unix", socketPath); err == nil {
		t.Fatal("expected socket to be removed")
	}
}
