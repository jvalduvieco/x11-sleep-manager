package main

import (
	"testing"
)

func TestBuildSessionRegistrationRequestReadsEnvironment(t *testing.T) {
	t.Setenv("DISPLAY", ":0")
	t.Setenv("XAUTHORITY", "/tmp/.Xauthority")
	t.Setenv("XDG_SESSION_TYPE", "x11")
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/tmp/bus")

	payload, err := buildSessionRegistrationRequest()
	if err != nil {
		t.Fatalf("buildSessionRegistrationRequest returned error: %v", err)
	}

	if payload.Display != ":0" {
		t.Fatalf("unexpected display: %q", payload.Display)
	}
	if payload.XAuthority != "/tmp/.Xauthority" {
		t.Fatalf("unexpected xauthority: %q", payload.XAuthority)
	}
	if payload.XDGSessionType != "x11" {
		t.Fatalf("unexpected session type: %q", payload.XDGSessionType)
	}
}

func TestBuildSessionRegistrationRequestRequiresDisplayAndXAuthority(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("XAUTHORITY", "")

	_, err := buildSessionRegistrationRequest()
	if err == nil {
		t.Fatal("expected missing env error")
	}
}

func TestReadConfigPatchInputFromArgument(t *testing.T) {
	body, err := readConfigPatchInput([]string{"{\"match\":{\"uid\":\"self\",\"who\":[\"Editor\"],\"what_any\":[\"idle\"]}}"})
	if err != nil {
		t.Fatalf("readConfigPatchInput returned error: %v", err)
	}
	if string(body) == "" {
		t.Fatal("expected non-empty patch body")
	}
}
