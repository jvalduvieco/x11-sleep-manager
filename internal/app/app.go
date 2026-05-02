package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/control"
	"github.com/jvalduvieco/x11_sleep_manager/internal/matcher"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

type App struct {
	server        *control.Server
	store         *state.Store
	config        config.Config
	source        observe.Source
	selfUID       int
	stopReconcile context.CancelFunc
}

func New(cfg config.Config, version string) *App {
	return NewWithSource(cfg, version, observe.NoopSource{}, os.Getuid())
}

func NewWithSource(cfg config.Config, version string, source observe.Source, selfUID int) *App {
	store := state.NewStore(cfg, version)
	return &App{
		server:  control.NewServerWithSessionRegistration(cfg.Socket.Path, store, store),
		store:   store,
		config:  cfg,
		source:  source,
		selfUID: selfUID,
	}
}

func (a *App) Start() error {
	if err := a.server.Start(); err != nil {
		a.store.SetLastError(err)
		a.store.Transition(state.ModeDegraded)
		return fmt.Errorf("start control server: %w", err)
	}

	reconcileCtx, cancel := context.WithCancel(context.Background())
	a.stopReconcile = cancel
	go a.runReconcileLoop(reconcileCtx)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.stopReconcile != nil {
		a.stopReconcile()
	}
	return a.server.Shutdown(ctx)
}

func (a *App) Snapshot() state.Snapshot {
	return a.store.Snapshot()
}

func (a *App) Reconcile(ctx context.Context) error {
	inhibitors, err := a.source.List(ctx)
	if err != nil {
		a.store.RecordReconcileFailure(err)
		a.store.Transition(state.ModeDegraded)
		return fmt.Errorf("list inhibitors: %w", err)
	}

	matched := matcher.Filter(inhibitors, a.config.Match, a.selfUID)
	a.store.RecordReconcileSuccess(matched)
	if len(matched) > 0 {
		a.store.Transition(state.ModeInhibited)
	} else {
		a.store.Transition(state.ModeIdle)
	}

	return nil
}

func (a *App) runReconcileLoop(ctx context.Context) {
	_ = a.Reconcile(ctx)

	ticker := time.NewTicker(a.config.Reconcile.Interval.Duration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = a.Reconcile(ctx)
		}
	}
}
