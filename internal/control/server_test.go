package control

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

func TestStatusEndpointServesSnapshot(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "x11-sm.sock")
	store := state.NewStore(config.Default(), "test")
	store.SetSessionReady(true)

	server := NewServer(socketPath, store)
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
	if snapshot.Version != "test" {
		t.Fatalf("unexpected version: %s", snapshot.Version)
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
