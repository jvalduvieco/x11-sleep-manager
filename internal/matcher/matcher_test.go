package matcher

import (
	"testing"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
)

func TestMatchesDefaultPolicy(t *testing.T) {
	match := config.Default().Match
	selfUID := 1000

	inhibitor := observe.Inhibitor{What: "sleep:idle", Who: "OpenCode", UID: selfUID}
	if !Matches(inhibitor, match, selfUID) {
		t.Fatal("expected inhibitor to match default policy")
	}
}

func TestMatchesRejectsWrongWho(t *testing.T) {
	match := config.Default().Match
	inhibitor := observe.Inhibitor{What: "sleep:idle", Who: "OtherApp", UID: 1000}

	if Matches(inhibitor, match, 1000) {
		t.Fatal("expected inhibitor to be rejected by who filter")
	}
}

func TestMatchesRejectsWrongUID(t *testing.T) {
	match := config.Default().Match
	inhibitor := observe.Inhibitor{What: "sleep:idle", Who: "OpenCode", UID: 1001}

	if Matches(inhibitor, match, 1000) {
		t.Fatal("expected inhibitor to be rejected by uid filter")
	}
}

func TestFilterReturnsOnlyMatchingInhibitors(t *testing.T) {
	match := config.Default().Match
	inhibitors := []observe.Inhibitor{
		{What: "sleep", Who: "OpenCode", UID: 1000},
		{What: "shutdown", Who: "OpenCode", UID: 1000},
		{What: "idle", Who: "OtherApp", UID: 1000},
	}

	matched := Filter(inhibitors, match, 1000)
	if got, want := len(matched), 1; got != want {
		t.Fatalf("matched length mismatch: got %d want %d", got, want)
	}
	if matched[0].What != "sleep" {
		t.Fatalf("unexpected matched inhibitor: %+v", matched[0])
	}
}
