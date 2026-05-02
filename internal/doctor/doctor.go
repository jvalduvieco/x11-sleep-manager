package doctor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
	"github.com/jvalduvieco/x11_sleep_manager/internal/x11state"
)

type Check struct {
	Name   string
	OK     bool
	Detail string
}

type Report struct {
	Checks []Check
}

type Runner struct {
	SocketCheck     func(ctx context.Context, socketPath string) error
	SystemBusCheck  func(ctx context.Context) error
	LogindCheck     func(ctx context.Context) error
	DisplayCheck    func() error
	XAuthorityCheck func() error
	XSetPathCheck   func() error
	XSetQueryCheck  func(ctx context.Context) error
}

func DefaultRunner() Runner {
	return Runner{
		SocketCheck:     checkSocket,
		SystemBusCheck:  checkSystemBus,
		LogindCheck:     checkLogind,
		DisplayCheck:    checkDisplay,
		XAuthorityCheck: checkXAuthority,
		XSetPathCheck:   checkXSetPath,
		XSetQueryCheck:  checkXSetQuery,
	}
}

func Run(ctx context.Context, socketPath string, runner Runner) Report {
	checks := []Check{
		runCheck(ctx, "socket", func(ctx context.Context) error { return runner.SocketCheck(ctx, socketPath) }),
		runCheck(ctx, "system-bus", runner.SystemBusCheck),
		runCheck(ctx, "logind", runner.LogindCheck),
		checkFromError("display", runner.DisplayCheck()),
		checkFromError("xauthority", runner.XAuthorityCheck()),
		checkFromError("xset-path", runner.XSetPathCheck()),
		runCheck(ctx, "xset-query", runner.XSetQueryCheck),
	}
	return Report{Checks: checks}
}

func (r Report) RenderText() string {
	var lines []string
	for _, check := range r.Checks {
		status := "FAIL"
		if check.OK {
			status = "PASS"
		}
		line := fmt.Sprintf("[%s] %s", status, check.Name)
		if check.Detail != "" {
			line += ": " + check.Detail
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (r Report) OK() bool {
	for _, check := range r.Checks {
		if !check.OK {
			return false
		}
	}
	return true
}

func runCheck(ctx context.Context, name string, fn func(context.Context) error) Check {
	return checkFromError(name, fn(ctx))
}

func checkFromError(name string, err error) Check {
	if err != nil {
		return Check{Name: name, OK: false, Detail: err.Error()}
	}
	return Check{Name: name, OK: true}
}

func checkSocket(ctx context.Context, socketPath string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
	resp, err := client.Get("http://unix/v1/status")
	if err != nil {
		return fmt.Errorf("request status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return nil
}

func checkSystemBus(context.Context) error {
	_, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("connect system bus: %w", err)
	}
	return nil
}

func checkLogind(ctx context.Context) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("connect system bus: %w", err)
	}
	call := conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").CallWithContext(ctx, "org.freedesktop.DBus.Peer.Ping", 0)
	if call.Err != nil {
		return fmt.Errorf("ping logind: %w", call.Err)
	}
	return nil
}

func checkDisplay() error {
	if os.Getenv("DISPLAY") == "" {
		return fmt.Errorf("DISPLAY is not set")
	}
	return nil
}

func checkXAuthority() error {
	if os.Getenv("XAUTHORITY") == "" {
		return fmt.Errorf("XAUTHORITY is not set")
	}
	return nil
}

func checkXSetPath() error {
	if _, err := exec.LookPath("xset"); err != nil {
		return fmt.Errorf("find xset: %w", err)
	}
	return nil
}

func checkXSetQuery(ctx context.Context) error {
	session := state.Session{
		Display:            os.Getenv("DISPLAY"),
		XAuthority:         os.Getenv("XAUTHORITY"),
		XDGSessionType:     os.Getenv("XDG_SESSION_TYPE"),
		DBusSessionBusAddr: os.Getenv("DBUS_SESSION_BUS_ADDRESS"),
	}
	output, err := (x11state.CommandRunner{}).Run(ctx, session, "q")
	if err != nil {
		return fmt.Errorf("run xset q: %w", err)
	}
	if _, err := x11state.ParseQuery(output); err != nil {
		return fmt.Errorf("parse xset q: %w", err)
	}
	return nil
}
