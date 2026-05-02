package state

import (
	"sync"
	"time"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/observe"
)

type Mode string

const (
	ModeDisabled  Mode = "disabled"
	ModeIdle      Mode = "idle"
	ModeInhibited Mode = "inhibited"
	ModeDegraded  Mode = "degraded"
)

type Snapshot struct {
	State                  Mode                `json:"state"`
	StartedAt              time.Time           `json:"started_at"`
	LastTransitionAt       time.Time           `json:"last_transition_at"`
	Config                 config.Config       `json:"config"`
	Version                string              `json:"version"`
	Session                *Session            `json:"session,omitempty"`
	SessionReady           bool                `json:"session_ready"`
	MatchingInhibitors     []observe.Inhibitor `json:"matching_inhibitors,omitempty"`
	MatchingInhibitorCount int                 `json:"matching_inhibitor_count"`
	LastReconcileAt        *time.Time          `json:"last_reconcile_at,omitempty"`
	ReconcileCount         int                 `json:"reconcile_count"`
	ReconcileFailures      int                 `json:"reconcile_failures"`
	X11OverridesActive     bool                `json:"x11_overrides_active"`
	LastError              string              `json:"last_error,omitempty"`
}

type Session struct {
	Display            string    `json:"display"`
	XAuthority         string    `json:"xauthority"`
	XDGSessionType     string    `json:"xdg_session_type,omitempty"`
	DBusSessionBusAddr string    `json:"dbus_session_bus_address,omitempty"`
	RegisteredAt       time.Time `json:"registered_at"`
}

type Store struct {
	mu                 sync.RWMutex
	state              Mode
	started            time.Time
	changed            time.Time
	config             config.Config
	version            string
	lastErr            string
	session            *Session
	matching           []observe.Inhibitor
	lastReconcileAt    *time.Time
	reconcileCount     int
	reconcileFailures  int
	x11OverridesActive bool
	clockNow           func() time.Time
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
		State:                  s.state,
		StartedAt:              s.started,
		LastTransitionAt:       s.changed,
		Config:                 s.config,
		Version:                s.version,
		Session:                cloneSession(s.session),
		SessionReady:           s.session != nil,
		MatchingInhibitors:     cloneInhibitors(s.matching),
		MatchingInhibitorCount: len(s.matching),
		LastReconcileAt:        cloneTimePtr(s.lastReconcileAt),
		ReconcileCount:         s.reconcileCount,
		ReconcileFailures:      s.reconcileFailures,
		X11OverridesActive:     s.x11OverridesActive,
		LastError:              s.lastErr,
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

func (s *Store) RegisterSession(session Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session.RegisteredAt = s.clockNow()
	s.session = &session
}

func (s *Store) RecordReconcileSuccess(matching []observe.Inhibitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clockNow()
	s.lastReconcileAt = &now
	s.reconcileCount++
	s.matching = cloneInhibitors(matching)
	s.lastErr = ""
}

func (s *Store) RecordReconcileFailure(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clockNow()
	s.lastReconcileAt = &now
	s.reconcileCount++
	s.reconcileFailures++
	if err != nil {
		s.lastErr = err.Error()
	}
}

func (s *Store) SetX11OverridesActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.x11OverridesActive = active
}

func cloneSession(session *Session) *Session {
	if session == nil {
		return nil
	}

	copy := *session
	return &copy
}

func cloneInhibitors(inhibitors []observe.Inhibitor) []observe.Inhibitor {
	if len(inhibitors) == 0 {
		return nil
	}
	cloned := make([]observe.Inhibitor, len(inhibitors))
	copy(cloned, inhibitors)
	return cloned
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
