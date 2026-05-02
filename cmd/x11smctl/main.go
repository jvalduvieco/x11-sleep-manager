package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
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
		return fmt.Errorf("usage: x11smctl [--socket path] status")
	}

	switch flag.Arg(0) {
	case "status":
		return printStatus(socketPath)
	default:
		return fmt.Errorf("unknown command %q", flag.Arg(0))
	}
}

func printStatus(socketPath string) error {
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
		return fmt.Errorf("status request failed: %s", resp.Status)
	}

	var payload any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode status response: %w", err)
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("format status response: %w", err)
	}

	fmt.Println(string(encoded))
	return nil
}
