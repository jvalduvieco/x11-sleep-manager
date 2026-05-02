package processctl

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"syscall"
	"testing"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
)

type fakeInspector struct {
	processes map[string][]int
	findErr   error
	signalErr error
	signals   []signalCall
}

type signalCall struct {
	pid    int
	signal syscall.Signal
}

func (f *fakeInspector) FindByName(_ context.Context, name string) ([]int, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return append([]int(nil), f.processes[name]...), nil
}

func (f *fakeInspector) Signal(_ context.Context, pid int, signal syscall.Signal) error {
	if f.signalErr != nil {
		return f.signalErr
	}
	f.signals = append(f.signals, signalCall{pid: pid, signal: signal})
	return nil
}

func TestPauseStopsConfiguredProcesses(t *testing.T) {
	inspector := &fakeInspector{processes: map[string][]int{"xss-lock": {100, 200}}}
	controller := NewController(config.Default().Processes, inspector)

	if err := controller.Pause(context.Background()); err != nil {
		t.Fatalf("Pause returned error: %v", err)
	}

	want := []signalCall{{pid: 100, signal: syscall.SIGSTOP}, {pid: 200, signal: syscall.SIGSTOP}}
	if !reflect.DeepEqual(inspector.signals, want) {
		t.Fatalf("unexpected signals: got %#v want %#v", inspector.signals, want)
	}
	if got, want := controller.PausedNames(), []string{"xss-lock"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected paused names: got %#v want %#v", got, want)
	}
}

func TestPauseIsIdempotentForAlreadyPausedProcesses(t *testing.T) {
	inspector := &fakeInspector{processes: map[string][]int{"xss-lock": {100}}}
	controller := NewController(config.Default().Processes, inspector)

	if err := controller.Pause(context.Background()); err != nil {
		t.Fatalf("Pause returned error: %v", err)
	}
	if err := controller.Pause(context.Background()); err != nil {
		t.Fatalf("second Pause returned error: %v", err)
	}
	if got, want := len(inspector.signals), 1; got != want {
		t.Fatalf("unexpected signal count: got %d want %d", got, want)
	}
}

func TestResumeOnlyResumesProcessesPausedByUs(t *testing.T) {
	inspector := &fakeInspector{processes: map[string][]int{"xss-lock": {100, 200}}}
	controller := NewController(config.Default().Processes, inspector)
	if err := controller.Pause(context.Background()); err != nil {
		t.Fatalf("Pause returned error: %v", err)
	}
	inspector.signals = nil

	if err := controller.Resume(context.Background()); err != nil {
		t.Fatalf("Resume returned error: %v", err)
	}

	sort.Slice(inspector.signals, func(i, j int) bool { return inspector.signals[i].pid < inspector.signals[j].pid })
	want := []signalCall{{pid: 100, signal: syscall.SIGCONT}, {pid: 200, signal: syscall.SIGCONT}}
	if !reflect.DeepEqual(inspector.signals, want) {
		t.Fatalf("unexpected signals: got %#v want %#v", inspector.signals, want)
	}
	if controller.HasPaused() {
		t.Fatal("expected paused set to be cleared")
	}
}

func TestPausePropagatesLookupErrors(t *testing.T) {
	controller := NewController(config.Default().Processes, &fakeInspector{findErr: errors.New("boom")})
	if err := controller.Pause(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
