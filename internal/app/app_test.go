package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

type fakeSource struct {
	inhibitors []observe.Inhibitor
	err        error
}

func (f fakeSource) List(context.Context) ([]observe.Inhibitor, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.inhibitors, nil
}

func TestReconcileMovesToInhibitedWhenMatchesExist(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "sleep:idle", Who: "OpenCode", UID: 1000}}}, 1000)

	if err := app.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile returned error: %v", err)
	}

	snapshot := app.Snapshot()
	if snapshot.State != state.ModeInhibited {
		t.Fatalf("unexpected state: %s", snapshot.State)
	}
	if snapshot.MatchingInhibitorCount != 1 {
		t.Fatalf("unexpected inhibitor count: %d", snapshot.MatchingInhibitorCount)
	}
}

func TestReconcileMovesToIdleWhenNoMatchesExist(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "shutdown", Who: "OpenCode", UID: 1000}}}, 1000)
	app.store.Transition(state.ModeInhibited)

	if err := app.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile returned error: %v", err)
	}

	if got := app.Snapshot().State; got != state.ModeIdle {
		t.Fatalf("unexpected state: %s", got)
	}
}

func TestReconcileMovesToDegradedOnSourceError(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{err: errors.New("boom")}, 1000)

	err := app.Reconcile(context.Background())
	if err == nil {
		t.Fatal("expected reconcile error")
	}

	snapshot := app.Snapshot()
	if snapshot.State != state.ModeDegraded {
		t.Fatalf("unexpected state: %s", snapshot.State)
	}
	if snapshot.ReconcileFailures != 1 {
		t.Fatalf("unexpected reconcile failures: %d", snapshot.ReconcileFailures)
	}
}

func TestReconcileLoopRunsImmediatelyOnStart(t *testing.T) {
	cfg := config.Default()
	cfg.Reconcile.Interval = config.Duration{Duration: time.Hour}
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "idle", Who: "OpenCode", UID: 1000}}}, 1000)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go app.runReconcileLoop(ctx)

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if app.Snapshot().ReconcileCount > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("expected initial reconcile to run")
}
