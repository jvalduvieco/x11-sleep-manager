package x11state

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/jvalduvieco/x11_sleep_manager/internal/config"
	"github.com/jvalduvieco/x11_sleep_manager/internal/state"
)

var (
	screenSaverPattern = regexp.MustCompile(`timeout:\s*(\d+)\s+cycle:\s*(\d+)`)
	dpmsPattern        = regexp.MustCompile(`DPMS is (Enabled|Disabled)`)
)

type Snapshot struct {
	ScreenSaverTimeout int
	ScreenSaverCycle   int
	DPMSEnabled        bool
}

type Runner interface {
	Run(ctx context.Context, session state.Session, args ...string) (string, error)
}

type Controller struct {
	runner  Runner
	config  config.X11Config
	saved   *Snapshot
	applied bool
}

type CommandRunner struct{}

func NewController(cfg config.X11Config, runner Runner) *Controller {
	return &Controller{config: cfg, runner: runner}
}

func (c *Controller) Apply(ctx context.Context, session state.Session) error {
	if c.applied {
		return nil
	}

	output, err := c.runner.Run(ctx, session, "q")
	if err != nil {
		return fmt.Errorf("query xset state: %w", err)
	}

	snapshot, err := ParseQuery(output)
	if err != nil {
		return fmt.Errorf("parse xset q output: %w", err)
	}

	if c.config.DisableScreenSaver {
		if _, err := c.runner.Run(ctx, session, "s", "off"); err != nil {
			return fmt.Errorf("disable screensaver: %w", err)
		}
	}
	if c.config.DisableDPMS {
		if _, err := c.runner.Run(ctx, session, "-dpms"); err != nil {
			return fmt.Errorf("disable dpms: %w", err)
		}
	}

	c.saved = &snapshot
	c.applied = true
	return nil
}

func (c *Controller) Restore(ctx context.Context, session state.Session) error {
	if !c.applied {
		return nil
	}
	if !c.config.RestorePreviousState || c.saved == nil {
		c.saved = nil
		c.applied = false
		return nil
	}

	if c.config.DisableScreenSaver {
		if _, err := c.runner.Run(ctx, session, "s", strconv.Itoa(c.saved.ScreenSaverTimeout), strconv.Itoa(c.saved.ScreenSaverCycle)); err != nil {
			return fmt.Errorf("restore screensaver: %w", err)
		}
	}
	if c.config.DisableDPMS {
		arg := "+dpms"
		if !c.saved.DPMSEnabled {
			arg = "-dpms"
		}
		if _, err := c.runner.Run(ctx, session, arg); err != nil {
			return fmt.Errorf("restore dpms: %w", err)
		}
	}

	c.saved = nil
	c.applied = false
	return nil
}

func (c *Controller) Applied() bool {
	return c.applied
}

func (c *Controller) SetConfig(cfg config.X11Config) {
	c.config = cfg
}

func ParseQuery(output string) (Snapshot, error) {
	screenSaverMatch := screenSaverPattern.FindStringSubmatch(output)
	if len(screenSaverMatch) != 3 {
		return Snapshot{}, fmt.Errorf("screen saver settings not found")
	}
	dpmsMatch := dpmsPattern.FindStringSubmatch(output)
	if len(dpmsMatch) != 2 {
		return Snapshot{}, fmt.Errorf("dpms status not found")
	}

	timeout, err := strconv.Atoi(screenSaverMatch[1])
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse screensaver timeout: %w", err)
	}
	cycle, err := strconv.Atoi(screenSaverMatch[2])
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse screensaver cycle: %w", err)
	}

	return Snapshot{
		ScreenSaverTimeout: timeout,
		ScreenSaverCycle:   cycle,
		DPMSEnabled:        dpmsMatch[1] == "Enabled",
	}, nil
}

func (CommandRunner) Run(ctx context.Context, session state.Session, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "xset", args...)
	cmd.Env = commandEnv(session)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("xset %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}

func commandEnv(session state.Session) []string {
	env := os.Environ()
	env = append(env, "DISPLAY="+session.Display)
	env = append(env, "XAUTHORITY="+session.XAuthority)
	if session.XDGSessionType != "" {
		env = append(env, "XDG_SESSION_TYPE="+session.XDGSessionType)
	}
	if session.DBusSessionBusAddr != "" {
		env = append(env, "DBUS_SESSION_BUS_ADDRESS="+session.DBusSessionBusAddr)
	}
	return env
}
