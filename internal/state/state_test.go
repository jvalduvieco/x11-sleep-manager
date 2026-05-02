package state

import (
	"errors"
	"testing"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
)

func TestNewStoreStartsIdle(t *testing.T) {
	store := NewStore(config.Default(), "dev")

	snapshot := store.Snapshot()
	if snapshot.State != ModeIdle {
		t.Fatalf("unexpected initial state: %s", snapshot.State)
	}
	if snapshot.Version != "dev" {
		t.Fatalf("unexpected version: %s", snapshot.Version)
	}
}

func TestTransitionUpdatesChangedAtOnlyOnChange(t *testing.T) {
	store := NewStore(config.Default(), "dev")
	initial := store.Snapshot().LastTransitionAt
	second := initial.Add(5 * time.Second)
	third := second.Add(5 * time.Second)

	store.clockNow = func() time.Time { return second }
	store.Transition(ModeIdle)
	if got := store.Snapshot().LastTransitionAt; !got.Equal(initial) {
		t.Fatalf("last transition changed unexpectedly: %s", got)
	}

	store.clockNow = func() time.Time { return third }
	store.Transition(ModeInhibited)
	if got := store.Snapshot().LastTransitionAt; !got.Equal(third) {
		t.Fatalf("last transition mismatch: got %s want %s", got, third)
	}
}

func TestSettersAreReflectedInSnapshot(t *testing.T) {
	store := NewStore(config.Default(), "dev")
	store.SetSessionReady(true)
	store.SetLastError(errors.New("boom"))

	snapshot := store.Snapshot()
	if !snapshot.SessionReady {
		t.Fatal("expected session to be ready")
	}
	if snapshot.LastError != "boom" {
		t.Fatalf("unexpected last error: %q", snapshot.LastError)
	}

	store.SetLastError(nil)
	if got := store.Snapshot().LastError; got != "" {
		t.Fatalf("expected cleared error, got %q", got)
	}
}
