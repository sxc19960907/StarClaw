package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCloudLifecycleStartStopStatus(t *testing.T) {
	var starts atomic.Int32
	controller := NewCloudLifecycleController(context.Background(), func(ctx context.Context) error {
		starts.Add(1)
		<-ctx.Done()
		return nil
	})

	controller.Start(context.Background())
	controller.Start(context.Background())
	waitForCloudLifecycleStarts(t, &starts, 1)
	if starts.Load() != 1 {
		t.Fatalf("starts = %d, want 1", starts.Load())
	}
	status := controller.Status()
	if !status.Running || status.StartedAt == nil || status.Configured || status.Enabled {
		t.Fatalf("status after start = %#v", status)
	}
	controller.Stop()
	waitCloudLifecycleStopped(t, controller)
	status = controller.Status()
	if status.Running || status.StoppedAt == nil {
		t.Fatalf("status after stop = %#v", status)
	}
	controller.Stop()
}

func waitForCloudLifecycleStarts(t *testing.T, starts *atomic.Int32, want int32) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if starts.Load() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("starts = %d, want at least %d", starts.Load(), want)
}

func TestCloudLifecycleRestartWaitsAndStartsFresh(t *testing.T) {
	var starts atomic.Int32
	controller := NewCloudLifecycleController(context.Background(), func(ctx context.Context) error {
		starts.Add(1)
		<-ctx.Done()
		return nil
	})

	controller.Start(context.Background())
	controller.Restart(context.Background())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if starts.Load() >= 2 && controller.Status().Running {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	status := controller.Status()
	if starts.Load() != 2 || !status.Running || status.RestartCount != 1 {
		t.Fatalf("starts=%d status=%#v, want restarted running controller", starts.Load(), status)
	}
	controller.Stop()
	waitCloudLifecycleStopped(t, controller)
}

func TestCloudLifecycleCapturesRunnerError(t *testing.T) {
	controller := NewCloudLifecycleController(context.Background(), func(context.Context) error {
		return errors.New("runner failed")
	})
	controller.Start(context.Background())
	waitCloudLifecycleStopped(t, controller)
	status := controller.Status()
	if status.Running || status.LastError != "runner failed" {
		t.Fatalf("status = %#v", status)
	}
}

func TestCloudLifecycleLocalOnlyState(t *testing.T) {
	controller := NewCloudLifecycleController(context.Background(), nil)
	controller.SetLocalOnlyState(true, false, "custom note")
	status := controller.Status()
	if !status.Configured || status.Enabled || status.Note != "custom note" {
		t.Fatalf("status = %#v", status)
	}
}

func waitCloudLifecycleStopped(t *testing.T, controller *CloudLifecycleController) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := controller.WaitStopped(ctx); err != nil {
		t.Fatalf("wait stopped: %v", err)
	}
}
