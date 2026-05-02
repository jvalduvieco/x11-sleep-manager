package processctl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
)

type Inspector interface {
	FindByName(ctx context.Context, name string) ([]int, error)
	Signal(ctx context.Context, pid int, signal syscall.Signal) error
}

type Controller struct {
	inspector Inspector
	config    config.ProcessesConfig
	paused    map[int]string
}

type ProcInspector struct{}

func NewController(cfg config.ProcessesConfig, inspector Inspector) *Controller {
	return &Controller{
		inspector: inspector,
		config:    cfg,
		paused:    make(map[int]string),
	}
}

func (c *Controller) Pause(ctx context.Context) error {
	for _, name := range c.config.Pause {
		pids, err := c.inspector.FindByName(ctx, name)
		if err != nil {
			return fmt.Errorf("find process %q: %w", name, err)
		}
		for _, pid := range pids {
			if _, exists := c.paused[pid]; exists {
				continue
			}
			if err := c.inspector.Signal(ctx, pid, syscall.SIGSTOP); err != nil {
				return fmt.Errorf("stop process %q (%d): %w", name, pid, err)
			}
			c.paused[pid] = name
		}
	}
	return nil
}

func (c *Controller) Resume(ctx context.Context) error {
	allowed := make(map[string]struct{}, len(c.config.Resume))
	for _, name := range c.config.Resume {
		allowed[name] = struct{}{}
	}

	for pid, name := range c.paused {
		if len(allowed) > 0 {
			if _, ok := allowed[name]; !ok {
				continue
			}
		}
		if err := c.inspector.Signal(ctx, pid, syscall.SIGCONT); err != nil {
			return fmt.Errorf("resume process %q (%d): %w", name, pid, err)
		}
		delete(c.paused, pid)
	}
	return nil
}

func (c *Controller) PausedNames() []string {
	seen := make(map[string]struct{}, len(c.paused))
	for _, name := range c.paused {
		seen[name] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Controller) HasPaused() bool {
	return len(c.paused) > 0
}

func (c *Controller) SetConfig(cfg config.ProcessesConfig) {
	c.config = cfg
}

func (ProcInspector) FindByName(_ context.Context, name string) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("read /proc: %w", err)
	}

	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		comm, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(comm)) == name {
			pids = append(pids, pid)
		}
	}
	sort.Ints(pids)
	return pids, nil
}

func (ProcInspector) Signal(_ context.Context, pid int, signal syscall.Signal) error {
	if err := syscall.Kill(pid, signal); err != nil {
		return fmt.Errorf("signal pid %d: %w", pid, err)
	}
	return nil
}
