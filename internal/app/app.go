package app

import (
	"context"
	"fmt"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/control"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

type App struct {
	server *control.Server
	store  *state.Store
	config config.Config
}

func New(cfg config.Config, version string) *App {
	store := state.NewStore(cfg, version)
	return &App{
		server: control.NewServerWithSessionRegistration(cfg.Socket.Path, store, store),
		store:  store,
		config: cfg,
	}
}

func (a *App) Start() error {
	if err := a.server.Start(); err != nil {
		a.store.SetLastError(err)
		a.store.Transition(state.ModeDegraded)
		return fmt.Errorf("start control server: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func (a *App) Snapshot() state.Snapshot {
	return a.store.Snapshot()
}
