package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/app"
	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to JSON config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	application := app.New(cfg, version)
	if err := application.Start(); err != nil {
		return err
	}

	log.Printf("x11-sleep-manager listening on %s", cfg.Socket.Path)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown daemon: %w", err)
	}

	return nil
}
