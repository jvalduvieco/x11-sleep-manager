package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
	"github.com/jvalduvieco/x11_sleep_manager/internal/x11state"
)

const sampleXSetOutput = `Screen Saver:
  prefer blanking:  yes    allow exposures:  yes
  timeout:  180    cycle:  240
DPMS (Energy Star):
  Standby: 600    Suspend: 600    Off: 600
  DPMS is Enabled
`

type fakeSource struct {
	inhibitors []observe.Inhibitor
	err        error
}

type fakeX11Runner struct {
	calls   [][]string
	outputs []string
	errs    []error
}

func (f fakeSource) List(context.Context) ([]observe.Inhibitor, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.inhibitors, nil
}

func (f *fakeX11Runner) Run(_ context.Context, _ state.Session, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	index := len(f.calls) - 1
	var output string
	if index < len(f.outputs) {
		output = f.outputs[index]
	}
	var err error
	if index < len(f.errs) {
		err = f.errs[index]
	}
	return output, err
}

func TestReconcileMovesToInhibitedWhenMatchesExist(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "sleep:idle", Who: "OpenCode", UID: 1000}}}, 1000)
	runner := &fakeX11Runner{outputs: []string{sampleXSetOutput}}
	app.x11 = x11state.NewController(cfg.X11, runner)
	app.store.RegisterSession(state.Session{Display: ":0", XAuthority: "/tmp/auth"})

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
	if !snapshot.X11OverridesActive {
		t.Fatal("expected x11 overrides to be active")
	}
}

func TestReconcileMovesToIdleWhenNoMatchesExist(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "shutdown", Who: "OpenCode", UID: 1000}}}, 1000)
	app.store.Transition(state.ModeInhibited)
	runner := &fakeX11Runner{outputs: []string{sampleXSetOutput}}
	app.x11 = x11state.NewController(cfg.X11, runner)
	app.store.RegisterSession(state.Session{Display: ":0", XAuthority: "/tmp/auth"})
	if err := app.x11.Apply(context.Background(), *app.store.Snapshot().Session); err != nil {
		t.Fatalf("setup apply returned error: %v", err)
	}
	app.store.SetX11OverridesActive(true)

	if err := app.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile returned error: %v", err)
	}

	if got := app.Snapshot().State; got != state.ModeIdle {
		t.Fatalf("unexpected state: %s", got)
	}
	if app.Snapshot().X11OverridesActive {
		t.Fatal("expected x11 overrides to be inactive")
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

func TestReconcileDegradesWithoutRegisteredSessionWhenMatchExists(t *testing.T) {
	cfg := config.Default()
	app := NewWithSource(cfg, "test", fakeSource{inhibitors: []observe.Inhibitor{{What: "idle", Who: "OpenCode", UID: 1000}}}, 1000)
	app.x11 = x11state.NewController(cfg.X11, &fakeX11Runner{outputs: []string{sampleXSetOutput}})

	err := app.Reconcile(context.Background())
	if err == nil {
		t.Fatal("expected reconcile error")
	}
	if got := app.Snapshot().State; got != state.ModeDegraded {
		t.Fatalf("unexpected state: %s", got)
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
