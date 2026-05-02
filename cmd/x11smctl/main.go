package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/doctor"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var socketPath string
	flag.StringVar(&socketPath, "socket", config.Default().Socket.Path, "path to daemon unix socket")
	flag.Parse()

	if flag.NArg() == 0 {
		return fmt.Errorf("usage: x11smctl [--socket path] <status|events|register-session|reconcile|enable|disable|doctor|config get>")
	}

	switch flag.Arg(0) {
	case "status":
		return printStatus(socketPath)
	case "events":
		return printJSONGet(socketPath, "/v1/events", "events")
	case "register-session":
		return registerSession(socketPath)
	case "reconcile":
		return postNoContentCommand(socketPath, "/v1/reconcile", "reconciled")
	case "enable":
		return postNoContentCommand(socketPath, "/v1/enable", "enabled")
	case "disable":
		return postNoContentCommand(socketPath, "/v1/disable", "disabled")
	case "doctor":
		return runDoctor(socketPath)
	case "config":
		if flag.NArg() < 2 || flag.Arg(1) != "get" {
			return fmt.Errorf("usage: x11smctl config get")
		}
		return printConfig(socketPath)
	default:
		return fmt.Errorf("unknown command %q", flag.Arg(0))
	}
}

func printStatus(socketPath string) error {
	return printJSONGet(socketPath, "/v1/status", "status")
}

type sessionRegistrationRequest struct {
	Display            string `json:"display"`
	XAuthority         string `json:"xauthority"`
	XDGSessionType     string `json:"xdg_session_type,omitempty"`
	DBusSessionBusAddr string `json:"dbus_session_bus_address,omitempty"`
}

func registerSession(socketPath string) error {
	payload, err := buildSessionRegistrationRequest()
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode session registration: %w", err)
	}

	client := newHTTPClient(socketPath)
	resp, err := client.Post("http://unix/v1/session", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("session registration failed: %s", resp.Status)
	}

	fmt.Println("session registered")
	return nil
}

func postNoContentCommand(socketPath, path, success string) error {
	client := newHTTPClient(socketPath)
	resp, err := client.Post("http://unix"+path, "application/json", nil)
	if err != nil {
		return fmt.Errorf("post %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request %s failed: %s", path, resp.Status)
	}
	fmt.Println(success)
	return nil
}

func runDoctor(socketPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	report := doctor.Run(ctx, socketPath, doctor.DefaultRunner())
	fmt.Println(report.RenderText())
	if !report.OK() {
		return fmt.Errorf("doctor checks failed")
	}
	return nil
}

func printConfig(socketPath string) error {
	return printJSONGet(socketPath, "/v1/config", "config")
}

func printJSONGet(socketPath, path, name string) error {
	client := newHTTPClient(socketPath)
	resp, err := client.Get("http://unix" + path)
	if err != nil {
		return fmt.Errorf("request %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s request failed: %s", name, resp.Status)
	}
	var payload any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode %s response: %w", name, err)
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("format %s response: %w", name, err)
	}
	fmt.Println(string(encoded))
	return nil
}

func buildSessionRegistrationRequest() (sessionRegistrationRequest, error) {
	payload := sessionRegistrationRequest{
		Display:            os.Getenv("DISPLAY"),
		XAuthority:         os.Getenv("XAUTHORITY"),
		XDGSessionType:     os.Getenv("XDG_SESSION_TYPE"),
		DBusSessionBusAddr: os.Getenv("DBUS_SESSION_BUS_ADDRESS"),
	}

	if payload.Display == "" {
		return sessionRegistrationRequest{}, fmt.Errorf("DISPLAY is not set")
	}
	if payload.XAuthority == "" {
		return sessionRegistrationRequest{}, fmt.Errorf("XAUTHORITY is not set")
	}

	return payload, nil
}

func newHTTPClient(socketPath string) *http.Client {
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
			},
		},
	}
}
