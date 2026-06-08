package sync

import (
	"reflect"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/starclaw/starclaw/internal/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	v := viper.New()
	SetDefaults(v)
	cfg, err := LoadConfig(v)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Enabled {
		t.Fatal("Enabled = true, want false")
	}
	if cfg.DryRun {
		t.Fatal("DryRun = true, want false")
	}
	if cfg.Endpoint != "" {
		t.Fatalf("Endpoint = %q, want empty", cfg.Endpoint)
	}
	if cfg.BatchMaxSessions != 25 {
		t.Fatalf("BatchMaxSessions = %d, want 25", cfg.BatchMaxSessions)
	}
	if cfg.BatchMaxBytes != 5*1024*1024 {
		t.Fatalf("BatchMaxBytes = %d, want 5242880", cfg.BatchMaxBytes)
	}
	if cfg.SingleSessionMaxBytes != 4*1024*1024 {
		t.Fatalf("SingleSessionMaxBytes = %d, want 4194304", cfg.SingleSessionMaxBytes)
	}
	if cfg.DaemonInterval != 24*time.Hour {
		t.Fatalf("DaemonInterval = %v, want 24h", cfg.DaemonInterval)
	}
	if cfg.DaemonStartupDelay != time.Minute {
		t.Fatalf("DaemonStartupDelay = %v, want 1m", cfg.DaemonStartupDelay)
	}
	if cfg.FailedMaxAttemptsTransient != 5 {
		t.Fatalf("FailedMaxAttemptsTransient = %d, want 5", cfg.FailedMaxAttemptsTransient)
	}
	if cfg.LockTimeout != 30*time.Second {
		t.Fatalf("LockTimeout = %v, want 30s", cfg.LockTimeout)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	v := viper.New()
	SetDefaults(v)
	v.Set("sync.enabled", true)
	v.Set("sync.dry_run", true)
	v.Set("sync.endpoint", "http://127.0.0.1/sync")
	v.Set("sync.exclude_agents", []string{"helper"})
	v.Set("sync.exclude_sources", []string{"remote"})
	v.Set("sync.batch_max_sessions", 3)
	v.Set("sync.batch_max_bytes", 1024)
	v.Set("sync.single_session_max_bytes", 512)
	v.Set("sync.daemon_interval", "2h")
	v.Set("sync.daemon_startup_delay", "5s")
	v.Set("sync.failed_max_attempts_transient", 7)
	v.Set("sync.lock_timeout", "250ms")

	cfg, err := LoadConfig(v)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !cfg.Enabled || !cfg.DryRun {
		t.Fatalf("Enabled/DryRun = %v/%v, want true/true", cfg.Enabled, cfg.DryRun)
	}
	if cfg.Endpoint != "http://127.0.0.1/sync" {
		t.Fatalf("Endpoint = %q", cfg.Endpoint)
	}
	if !reflect.DeepEqual(cfg.ExcludeAgents, []string{"helper"}) {
		t.Fatalf("ExcludeAgents = %v", cfg.ExcludeAgents)
	}
	if !reflect.DeepEqual(cfg.ExcludeSources, []string{"remote"}) {
		t.Fatalf("ExcludeSources = %v", cfg.ExcludeSources)
	}
	if cfg.BatchMaxSessions != 3 || cfg.BatchMaxBytes != 1024 || cfg.SingleSessionMaxBytes != 512 {
		t.Fatalf("batch caps = %+v", cfg)
	}
	if cfg.DaemonInterval != 2*time.Hour || cfg.DaemonStartupDelay != 5*time.Second || cfg.LockTimeout != 250*time.Millisecond {
		t.Fatalf("durations = %+v", cfg)
	}
	if cfg.FailedMaxAttemptsTransient != 7 {
		t.Fatalf("FailedMaxAttemptsTransient = %d", cfg.FailedMaxAttemptsTransient)
	}
}

func TestFromConfigParsesDurations(t *testing.T) {
	cfg, err := FromConfig(config.SyncConfig{
		Enabled:                    true,
		DryRun:                     true,
		BatchMaxSessions:           2,
		BatchMaxBytes:              100,
		SingleSessionMaxBytes:      50,
		DaemonInterval:             "3h",
		DaemonStartupDelay:         "4s",
		FailedMaxAttemptsTransient: 6,
		LockTimeout:                "1s",
	})
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if !cfg.Enabled || !cfg.DryRun {
		t.Fatalf("Enabled/DryRun = %v/%v, want true/true", cfg.Enabled, cfg.DryRun)
	}
	if cfg.DaemonInterval != 3*time.Hour || cfg.DaemonStartupDelay != 4*time.Second || cfg.LockTimeout != time.Second {
		t.Fatalf("durations = %+v", cfg)
	}
}

func TestLoadConfigRejectsInvalidDuration(t *testing.T) {
	v := viper.New()
	SetDefaults(v)
	v.Set("sync.lock_timeout", "not-a-duration")
	if _, err := LoadConfig(v); err == nil {
		t.Fatal("LoadConfig should reject invalid duration")
	}
}
