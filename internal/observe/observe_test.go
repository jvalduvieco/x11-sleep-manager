package observe

import (
	"context"
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

type fakeManager struct {
	inhibitors []rawInhibitor
	err        error
}

func (f fakeManager) ListInhibitors(context.Context) ([]rawInhibitor, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.inhibitors, nil
}

type fakeObject struct {
	call *dbus.Call
}

func (f fakeObject) CallWithContext(context.Context, string, dbus.Flags, ...any) *dbus.Call {
	return f.call
}

func TestLogindSourceNormalizesInhibitors(t *testing.T) {
	source := NewLogindSourceWithManager(fakeManager{inhibitors: []rawInhibitor{{
		What: "sleep:idle",
		Who:  "OpenCode",
		Why:  "test",
		UID:  1000,
		PID:  42,
	}}})

	inhibitors, err := source.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if got, want := len(inhibitors), 1; got != want {
		t.Fatalf("unexpected inhibitor count: got %d want %d", got, want)
	}
	if inhibitors[0].UID != 1000 || inhibitors[0].What != "sleep:idle" {
		t.Fatalf("unexpected inhibitor: %+v", inhibitors[0])
	}
}

func TestLogindSourcePropagatesErrors(t *testing.T) {
	source := NewLogindSourceWithManager(fakeManager{err: errors.New("boom")})

	_, err := source.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLogindManagerWrapsCallErrors(t *testing.T) {
	manager := logindManager{object: fakeObject{call: &dbus.Call{Err: errors.New("dbus failed")}}}

	_, err := manager.ListInhibitors(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStaticErrorSourceReturnsConfiguredError(t *testing.T) {
	source := StaticErrorSource{Err: errors.New("startup failed")}

	_, err := source.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
