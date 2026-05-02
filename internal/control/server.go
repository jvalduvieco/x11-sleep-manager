package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

type StatusSource interface {
	Snapshot() state.Snapshot
}

type ConfigSource interface {
	Config() config.Config
}

type SessionRegistrar interface {
	RegisterSession(session state.Session)
}

type RuntimeController interface {
	Reconcile(ctx context.Context) error
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
}

type Server struct {
	httpServer *http.Server
	listener   net.Listener
	socketPath string
}

func NewServer(socketPath string, source StatusSource) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(source.Snapshot()); err != nil {
			http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		}
	})

	return &Server{
		httpServer: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		socketPath: socketPath,
	}
}

func NewServerWithSessionRegistration(socketPath string, source StatusSource, registrar SessionRegistrar) *Server {
	return NewServerWithRuntimeControl(socketPath, source, registrar, nil)
}

func NewServerWithRuntimeControl(socketPath string, source StatusSource, registrar SessionRegistrar, runtime RuntimeController) *Server {
	mux := http.NewServeMux()
	if configSource, ok := source.(ConfigSource); ok {
		mux.HandleFunc("GET /v1/config", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(configSource.Config()); err != nil {
				http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
			}
		})
	}
	mux.HandleFunc("GET /v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(source.Snapshot()); err != nil {
			http.Error(w, fmt.Sprintf("encode response: %v", err), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("POST /v1/session", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var payload state.Session
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, fmt.Sprintf("decode request: %v", err), http.StatusBadRequest)
			return
		}
		if payload.Display == "" {
			http.Error(w, "display must not be empty", http.StatusBadRequest)
			return
		}
		if payload.XAuthority == "" {
			http.Error(w, "xauthority must not be empty", http.StatusBadRequest)
			return
		}

		registrar.RegisterSession(payload)
		w.WriteHeader(http.StatusNoContent)
	})
	if runtime != nil {
		mux.HandleFunc("POST /v1/reconcile", func(w http.ResponseWriter, r *http.Request) {
			if err := runtime.Reconcile(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
		mux.HandleFunc("POST /v1/enable", func(w http.ResponseWriter, r *http.Request) {
			if err := runtime.Enable(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
		mux.HandleFunc("POST /v1/disable", func(w http.ResponseWriter, r *http.Request) {
			if err := runtime.Disable(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}

	return &Server{
		httpServer: &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
		socketPath: socketPath,
	}
}

func (s *Server) Start() error {
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0o755); err != nil {
		return fmt.Errorf("create socket directory: %w", err)
	}

	if err := os.RemoveAll(s.socketPath); err != nil {
		return fmt.Errorf("remove stale socket: %w", err)
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen on unix socket: %w", err)
	}

	if err := os.Chmod(s.socketPath, 0o600); err != nil {
		listener.Close()
		return fmt.Errorf("chmod socket: %w", err)
	}

	s.listener = listener

	go func() {
		_ = s.httpServer.Serve(listener)
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.listener == nil {
		return nil
	}

	err := s.httpServer.Shutdown(ctx)
	removeErr := os.Remove(s.socketPath)
	if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return errors.Join(err, fmt.Errorf("remove socket: %w", removeErr))
	}
	return err
}
