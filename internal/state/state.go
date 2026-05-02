package state

import (
	"sync"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
)

type Mode string

const (
	ModeDisabled  Mode = "disabled"
	ModeIdle      Mode = "idle"
	ModeInhibited Mode = "inhibited"
	ModeDegraded  Mode = "degraded"
)

type Snapshot struct {
	State            Mode          `json:"state"`
	StartedAt        time.Time     `json:"started_at"`
	LastTransitionAt time.Time     `json:"last_transition_at"`
	Config           config.Config `json:"config"`
	Version          string        `json:"version"`
	SessionReady     bool          `json:"session_ready"`
	LastError        string        `json:"last_error,omitempty"`
}

type Store struct {
	mu       sync.RWMutex
	state    Mode
	started  time.Time
	changed  time.Time
	config   config.Config
	version  string
	lastErr  string
	session  bool
	clockNow func() time.Time
}

func NewStore(cfg config.Config, version string) *Store {
	now := time.Now().UTC()
	return &Store{
		state:    ModeIdle,
		started:  now,
		changed:  now,
		config:   cfg,
		version:  version,
		clockNow: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		State:            s.state,
		StartedAt:        s.started,
		LastTransitionAt: s.changed,
		Config:           s.config,
		Version:          s.version,
		SessionReady:     s.session,
		LastError:        s.lastErr,
	}
}

func (s *Store) Transition(next Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == next {
		return
	}

	s.state = next
	s.changed = s.clockNow()
}

func (s *Store) SetLastError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil {
		s.lastErr = ""
		return
	}

	s.lastErr = err.Error()
}

func (s *Store) SetSessionReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session = ready
}
