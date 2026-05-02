package doctor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunCollectsChecks(t *testing.T) {
	report := Run(context.Background(), "/tmp/test.sock", Runner{
		SocketCheck:     func(context.Context, string) error { return nil },
		SystemBusCheck:  func(context.Context) error { return errors.New("bus") },
		LogindCheck:     func(context.Context) error { return nil },
		DisplayCheck:    func() error { return nil },
		XAuthorityCheck: func() error { return nil },
		XSetPathCheck:   func() error { return nil },
		XSetQueryCheck:  func(context.Context) error { return errors.New("xset") },
	})

	if got, want := len(report.Checks), 7; got != want {
		t.Fatalf("unexpected check count: got %d want %d", got, want)
	}
	if report.Checks[1].OK {
		t.Fatal("expected system-bus check to fail")
	}
	if report.Checks[6].OK {
		t.Fatal("expected xset-query check to fail")
	}
	if report.OK() {
		t.Fatal("expected report to be failing")
	}
}

func TestRenderTextIncludesPassAndFail(t *testing.T) {
	report := Report{Checks: []Check{{Name: "socket", OK: true}, {Name: "xset", OK: false, Detail: "boom"}}}
	output := report.RenderText()
	if !strings.Contains(output, "[PASS] socket") {
		t.Fatalf("missing pass line: %q", output)
	}
	if !strings.Contains(output, "[FAIL] xset: boom") {
		t.Fatalf("missing fail line: %q", output)
	}
	if report.OK() {
		t.Fatal("expected report to be failing")
	}
}
