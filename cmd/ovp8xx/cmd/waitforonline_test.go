package cmd

import (
	"context"
	"errors"
	"testing"
	"time"
)

// never is a check which is not successful until the context is done.
func never(ctx context.Context) (bool, error) {
	<-ctx.Done()
	return false, ctx.Err()
}

func TestWaitForAnyReturnsFirstMode(t *testing.T) {
	cancelled := make(chan struct{})
	checks := map[string]availabilityCheck{
		waitModeProductive: func(ctx context.Context) (bool, error) {
			<-ctx.Done()
			close(cancelled)
			return false, ctx.Err()
		},
		waitModeSWUpdate: func(context.Context) (bool, error) { return true, nil },
	}
	mode, err := waitForAny(context.Background(), checks)
	if err != nil || mode != waitModeSWUpdate {
		t.Fatalf("expected the swupdate mode, got %q, %v", mode, err)
	}
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Error("the remaining check was not cancelled")
	}
}

func TestWaitForAnyTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	mode, err := waitForAny(ctx, map[string]availabilityCheck{
		waitModeProductive: never,
		waitModeSWUpdate:   never,
	})
	if mode != "" || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected a timeout, got %q, %v", mode, err)
	}
}
