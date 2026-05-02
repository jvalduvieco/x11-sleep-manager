package x11state

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

const sampleXSetOutput = `Keyboard Control:
  auto repeat:  on    key click percent:  0    LED mask:  00000000
Screen Saver:
  prefer blanking:  yes    allow exposures:  yes
  timeout:  180    cycle:  240
DPMS (Energy Star):
  Standby: 600    Suspend: 600    Off: 600
  DPMS is Enabled
`

type fakeRunner struct {
	calls   [][]string
	outputs []string
	errs    []error
}

func (f *fakeRunner) Run(_ context.Context, _ state.Session, args ...string) (string, error) {
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

func TestParseQuery(t *testing.T) {
	snapshot, err := ParseQuery(sampleXSetOutput)
	if err != nil {
		t.Fatalf("ParseQuery returned error: %v", err)
	}
	if snapshot.ScreenSaverTimeout != 180 || snapshot.ScreenSaverCycle != 240 || !snapshot.DPMSEnabled {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestApplyDisablesScreenSaverAndDPMS(t *testing.T) {
	runner := &fakeRunner{outputs: []string{sampleXSetOutput}}
	controller := NewController(config.Default().X11, runner)
	session := state.Session{Display: ":0", XAuthority: "/tmp/auth"}

	if err := controller.Apply(context.Background(), session); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	want := [][]string{{"q"}, {"s", "off"}, {"-dpms"}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected calls: got %#v want %#v", runner.calls, want)
	}
	if !controller.Applied() {
		t.Fatal("expected controller to be marked applied")
	}
}

func TestApplyIsIdempotent(t *testing.T) {
	runner := &fakeRunner{outputs: []string{sampleXSetOutput}}
	controller := NewController(config.Default().X11, runner)
	session := state.Session{Display: ":0", XAuthority: "/tmp/auth"}

	if err := controller.Apply(context.Background(), session); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := controller.Apply(context.Background(), session); err != nil {
		t.Fatalf("second Apply returned error: %v", err)
	}
	if got, want := len(runner.calls), 3; got != want {
		t.Fatalf("unexpected call count: got %d want %d", got, want)
	}
}

func TestRestoreRestoresSavedState(t *testing.T) {
	runner := &fakeRunner{outputs: []string{sampleXSetOutput}}
	controller := NewController(config.Default().X11, runner)
	session := state.Session{Display: ":0", XAuthority: "/tmp/auth"}

	if err := controller.Apply(context.Background(), session); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if err := controller.Restore(context.Background(), session); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	want := [][]string{{"q"}, {"s", "off"}, {"-dpms"}, {"s", "180", "240"}, {"+dpms"}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("unexpected calls: got %#v want %#v", runner.calls, want)
	}
	if controller.Applied() {
		t.Fatal("expected controller to be cleared after restore")
	}
}

func TestRestoreWithoutApplyDoesNothing(t *testing.T) {
	runner := &fakeRunner{}
	controller := NewController(config.Default().X11, runner)
	session := state.Session{Display: ":0", XAuthority: "/tmp/auth"}

	if err := controller.Restore(context.Background(), session); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if len(runner.calls) != 0 {
		t.Fatalf("unexpected calls: %#v", runner.calls)
	}
}

func TestApplyPropagatesQueryErrors(t *testing.T) {
	runner := &fakeRunner{errs: []error{errors.New("boom")}}
	controller := NewController(config.Default().X11, runner)
	session := state.Session{Display: ":0", XAuthority: "/tmp/auth"}

	if err := controller.Apply(context.Background(), session); err == nil {
		t.Fatal("expected error")
	}
}
