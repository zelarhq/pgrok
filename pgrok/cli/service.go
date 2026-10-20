package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	charmlog "charm.land/log/v2"
	"github.com/kardianos/service"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"unknwon.dev/x/anyx"
	"unknwon.dev/x/logx"

	"github.com/pgrok/pgrok/internal/osutil"
)

const serviceName = "pgrok"

// commandService manages pgrok as a background service that starts on boot (or
// login on macOS) and is restarted by the OS service manager if it exits:
//
//   - macOS: per-user LaunchAgent, no root needed, logs in ~/Library/Logs
//   - Linux: systemd system unit running as the invoking user (sudo required),
//     so it survives logout; logs via journalctl
//   - Windows: Windows service (Administrator required), logs next to the config
//
// The service always runs `pgrok http` with the config file written by `pgrok
// init`.
func commandService(homeDir string, logger *logx.Logger) *cli.Command {
	action := func(name string, fn func(ctx context.Context, cmd *cli.Command, s service.Service, configPath string) error) *cli.Command {
		return &cli.Command{
			Name:  name,
			Usage: serviceUsages[name],
			Flags: commonFlags(homeDir, logger),
			Action: func(ctx context.Context, cmd *cli.Command) error {
				configPath, err := filepath.Abs(cmd.String("config"))
				if err != nil {
					return errors.Wrap(err, "resolve config path")
				}
				s, err := newService(configPath, nil)
				if err != nil {
					return err
				}
				return fn(ctx, cmd, s, configPath)
			},
		}
	}

	return &cli.Command{
		Name:  "service",
		Usage: "Run pgrok as a background service that starts automatically",
		Commands: []*cli.Command{
			action("install", serviceInstall(logger)),
			action("uninstall", func(_ context.Context, _ *cli.Command, s service.Service, _ string) error {
				_ = s.Stop()
				if err := s.Uninstall(); err != nil {
					return errors.Wrap(err, "uninstall service")
				}
				logger.Info("Service uninstalled")
				return nil
			}),
			action("start", serviceControl(logger, "start")),
			action("stop", serviceControl(logger, "stop")),
			action("restart", serviceControl(logger, "restart")),
			action("status", func(_ context.Context, _ *cli.Command, s service.Service, configPath string) error {
				status, err := s.Status()
				if err != nil {
					return errors.Wrap(err, "query service status")
				}
				state := map[service.Status]string{
					service.StatusRunning: "running",
					service.StatusStopped: "stopped",
				}[status]
				fmt.Printf("Status: %s\nConfig: %s\nLogs:   %s\n", anyx.Coalesce(state, "unknown"), configPath, serviceLogHint(configPath))
				return nil
			}),
			{
				// Invoked by the OS service manager, not by users.
				Name:   "run",
				Hidden: true,
				Flags:  commandHTTP(homeDir, logger).Flags,
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath, err := filepath.Abs(cmd.String("config"))
					if err != nil {
						return errors.Wrap(err, "resolve config path")
					}
					runLogger := logger
					if runtime.GOOS == "windows" {
						// Windows services have no stdout/stderr, so log next to the config.
						// ponytail: truncated on every start, add rotation if anyone needs history.
						f, err := os.Create(windowsLogPath(configPath))
						if err == nil {
							defer func() { _ = f.Close() }()
							runLogger = logx.New(charmlog.NewWithOptions(f, charmlog.Options{ReportTimestamp: true}))
						}
					}
					s, err := newService(configPath, &program{run: func() {
						_ = actionHTTP(ctx, cmd, runLogger)
					}})
					if err != nil {
						return err
					}
					return s.Run()
				},
			},
		},
	}
}

var serviceUsages = map[string]string{
	"install":   "Install and start the service using the config file from `pgrok init`",
	"uninstall": "Stop and remove the service",
	"start":     "Start the service",
	"stop":      "Stop the service",
	"restart":   "Restart the service (e.g. after editing the config file)",
	"status":    "Show whether the service is running and where it logs",
}

func serviceInstall(logger *logx.Logger) func(context.Context, *cli.Command, service.Service, string) error {
	return func(_ context.Context, _ *cli.Command, s service.Service, configPath string) error {
		if !osutil.IsExist(configPath) {
			return errors.Errorf("config file %q not found, run \"pgrok init\" first or pass --config", configPath)
		}
		if _, err := loadConfig(configPath); err != nil {
			return errors.Wrapf(err, "load config %q", configPath)
		}

		// Reinstalling picks up a new binary path or config path.
		if _, err := s.Status(); !errors.Is(err, service.ErrNotInstalled) {
			_ = s.Stop()
			_ = s.Uninstall()
		}
		if err := s.Install(); err != nil {
			return errors.Wrap(err, "install service"+privilegeHint())
		}
		if err := s.Start(); err != nil {
			return errors.Wrap(err, "start service")
		}
		logger.Info("Service installed and started",
			"config", configPath,
			"logs", serviceLogHint(configPath),
		)
		return nil
	}
}

func serviceControl(logger *logx.Logger, action string) func(context.Context, *cli.Command, service.Service, string) error {
	return func(_ context.Context, _ *cli.Command, s service.Service, _ string) error {
		if err := service.Control(s, action); err != nil {
			return errors.Wrapf(err, "%s service%s", action, privilegeHint())
		}
		logger.Info("Service " + action + " done")
		return nil
	}
}

func newService(configPath string, prg service.Interface) (service.Service, error) {
	if prg == nil {
		prg = &program{}
	}
	config := &service.Config{
		Name:        serviceName,
		DisplayName: "pgrok tunnel",
		Description: "Keeps a pgrok HTTP tunnel connected.",
		Arguments:   []string{"service", "run", "--config", configPath},
		Option:      service.KeyValue{},
	}
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		config.Option["UserService"] = true
		config.Option["RunAtLoad"] = true
		config.Option["LogOutput"] = true
		config.Option["LogDirectory"] = filepath.Join(home, "Library", "Logs")
	case "windows":
		config.Option["OnFailure"] = "restart"
		config.Option["OnFailureDelayDuration"] = "5s"
	default:
		// System unit so the tunnel survives logout, but run as the human user
		// who owns the config rather than root.
		config.UserName = os.Getenv("SUDO_USER")
		config.Option["SystemdScript"] = systemdUnit
	}

	s, err := service.New(prg, config)
	return s, errors.Wrap(err, "create service")
}

// systemdUnit replaces the library default, which waits 120s between restarts
// and gives up after 10 failures; a tunnel should come back within seconds and
// keep retrying while the network is down.
const systemdUnit = `[Unit]
Description={{Description}}
ConditionFileIsExecutable={{Path | cmdEscape}}
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=0

[Service]
ExecStart={{Path | cmdEscape}}{{range Arguments}} {{. | cmd}}{{end}}
{{if UserName}}User={{UserName}}
{{end}}Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`

func serviceLogHint(configPath string) string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", serviceName+".err.log")
	case "windows":
		return windowsLogPath(configPath)
	default:
		return "journalctl -u " + serviceName + " -f"
	}
}

func windowsLogPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "pgrok.log")
}

func privilegeHint() string {
	switch runtime.GOOS {
	case "windows":
		return " (run from an Administrator PowerShell)"
	case "linux":
		return " (run with sudo)"
	}
	return ""
}

// program adapts the blocking tunnel loop to the service lifecycle.
type program struct {
	run func()
}

func (p *program) Start(service.Service) error {
	if p.run != nil {
		go p.run()
	}
	return nil
}

func (*program) Stop(service.Service) error { return nil }
