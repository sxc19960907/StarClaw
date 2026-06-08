// Package sync provides local-only session sync foundations. It does not
// include a cloud uploader or daemon runner in this phase.
package sync

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/starclaw/starclaw/internal/config"
)

// Config is the runtime sync config with parsed durations.
type Config struct {
	Enabled                    bool
	DryRun                     bool
	Endpoint                   string
	ExcludeAgents              []string
	ExcludeSources             []string
	BatchMaxSessions           int
	BatchMaxBytes              int
	SingleSessionMaxBytes      int
	DaemonInterval             time.Duration
	DaemonStartupDelay         time.Duration
	FailedMaxAttemptsTransient int
	LockTimeout                time.Duration
}

// SetDefaults registers sync.* defaults on a viper instance.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("sync.enabled", false)
	v.SetDefault("sync.dry_run", false)
	v.SetDefault("sync.endpoint", "")
	v.SetDefault("sync.exclude_agents", []string{})
	v.SetDefault("sync.exclude_sources", []string{})
	v.SetDefault("sync.batch_max_sessions", 25)
	v.SetDefault("sync.batch_max_bytes", 5*1024*1024)
	v.SetDefault("sync.single_session_max_bytes", 4*1024*1024)
	v.SetDefault("sync.daemon_interval", "24h")
	v.SetDefault("sync.daemon_startup_delay", "60s")
	v.SetDefault("sync.failed_max_attempts_transient", 5)
	v.SetDefault("sync.lock_timeout", "30s")
}

// LoadConfig loads runtime sync config from viper.
func LoadConfig(v *viper.Viper) (Config, error) {
	daemonInterval, err := parseDuration(v.GetString("sync.daemon_interval"), "sync.daemon_interval")
	if err != nil {
		return Config{}, err
	}
	startupDelay, err := parseDuration(v.GetString("sync.daemon_startup_delay"), "sync.daemon_startup_delay")
	if err != nil {
		return Config{}, err
	}
	lockTimeout, err := parseDuration(v.GetString("sync.lock_timeout"), "sync.lock_timeout")
	if err != nil {
		return Config{}, err
	}
	return Config{
		Enabled:                    v.GetBool("sync.enabled"),
		DryRun:                     v.GetBool("sync.dry_run"),
		Endpoint:                   v.GetString("sync.endpoint"),
		ExcludeAgents:              v.GetStringSlice("sync.exclude_agents"),
		ExcludeSources:             v.GetStringSlice("sync.exclude_sources"),
		BatchMaxSessions:           v.GetInt("sync.batch_max_sessions"),
		BatchMaxBytes:              v.GetInt("sync.batch_max_bytes"),
		SingleSessionMaxBytes:      v.GetInt("sync.single_session_max_bytes"),
		DaemonInterval:             daemonInterval,
		DaemonStartupDelay:         startupDelay,
		FailedMaxAttemptsTransient: v.GetInt("sync.failed_max_attempts_transient"),
		LockTimeout:                lockTimeout,
	}, nil
}

// FromConfig converts config.SyncConfig into runtime sync config.
func FromConfig(cfg config.SyncConfig) (Config, error) {
	daemonInterval, err := parseDuration(cfg.DaemonInterval, "sync.daemon_interval")
	if err != nil {
		return Config{}, err
	}
	startupDelay, err := parseDuration(cfg.DaemonStartupDelay, "sync.daemon_startup_delay")
	if err != nil {
		return Config{}, err
	}
	lockTimeout, err := parseDuration(cfg.LockTimeout, "sync.lock_timeout")
	if err != nil {
		return Config{}, err
	}
	return Config{
		Enabled:                    cfg.Enabled,
		DryRun:                     cfg.DryRun,
		Endpoint:                   cfg.Endpoint,
		ExcludeAgents:              append([]string(nil), cfg.ExcludeAgents...),
		ExcludeSources:             append([]string(nil), cfg.ExcludeSources...),
		BatchMaxSessions:           cfg.BatchMaxSessions,
		BatchMaxBytes:              cfg.BatchMaxBytes,
		SingleSessionMaxBytes:      cfg.SingleSessionMaxBytes,
		DaemonInterval:             daemonInterval,
		DaemonStartupDelay:         startupDelay,
		FailedMaxAttemptsTransient: cfg.FailedMaxAttemptsTransient,
		LockTimeout:                lockTimeout,
	}, nil
}

func parseDuration(value, field string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", field, err)
	}
	return d, nil
}
