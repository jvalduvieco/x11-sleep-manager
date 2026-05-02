package observe

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	logindService = "org.freedesktop.login1"
	logindPath    = "/org/freedesktop/login1"
	logindMethod  = "org.freedesktop.login1.Manager.ListInhibitors"
)

type Inhibitor struct {
	What string `json:"what"`
	Who  string `json:"who"`
	Why  string `json:"why,omitempty"`
	UID  int    `json:"uid"`
}

type Source interface {
	List(ctx context.Context) ([]Inhibitor, error)
}

type NoopSource struct{}

func (NoopSource) List(context.Context) ([]Inhibitor, error) {
	return nil, nil
}

type StaticErrorSource struct {
	Err error
}

func (s StaticErrorSource) List(context.Context) ([]Inhibitor, error) {
	if s.Err == nil {
		return nil, nil
	}
	return nil, s.Err
}

type LogindSource struct {
	manager inhibitorLister
}

type inhibitorLister interface {
	ListInhibitors(ctx context.Context) ([]rawInhibitor, error)
}

type rawInhibitor struct {
	What string
	Who  string
	Why  string
	Mode string
	UID  uint32
	PID  uint32
}

type dbusObject interface {
	CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...any) *dbus.Call
}

type logindManager struct {
	object dbusObject
}

func NewSystemBusSource() (Source, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect system bus: %w", err)
	}

	return NewLogindSource(conn), nil
}

func NewLogindSource(conn *dbus.Conn) *LogindSource {
	return &LogindSource{
		manager: logindManager{
			object: conn.Object(logindService, dbus.ObjectPath(logindPath)),
		},
	}
}

func NewLogindSourceWithManager(manager inhibitorLister) *LogindSource {
	return &LogindSource{manager: manager}
}

func (s *LogindSource) List(ctx context.Context) ([]Inhibitor, error) {
	raw, err := s.manager.ListInhibitors(ctx)
	if err != nil {
		return nil, err
	}

	inhibitors := make([]Inhibitor, 0, len(raw))
	for _, item := range raw {
		inhibitors = append(inhibitors, Inhibitor{
			What: item.What,
			Who:  item.Who,
			Why:  item.Why,
			UID:  int(item.UID),
		})
	}
	return inhibitors, nil
}

func (m logindManager) ListInhibitors(ctx context.Context) ([]rawInhibitor, error) {
	call := m.object.CallWithContext(ctx, logindMethod, 0)
	if call.Err != nil {
		return nil, fmt.Errorf("call %s: %w", logindMethod, call.Err)
	}

	var raw []rawInhibitor
	if err := call.Store(&raw); err != nil {
		return nil, fmt.Errorf("decode inhibitors: %w", err)
	}
	return raw, nil
}
