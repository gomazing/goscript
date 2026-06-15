package goscript

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestSchedulerSubmitHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	scheduler := NewScheduler(1)
	scheduler.Start(ctx)

	blocker := make(chan struct{})
	defer func() {
		close(blocker)
		scheduler.Stop()
	}()

	if err := scheduler.Submit(Task{
		Name: "blocker",
		Handler: func(ctx context.Context) (interface{}, error) {
			<-blocker
			return "blocked", nil
		},
	}); err != nil {
		t.Fatalf("unexpected blocker submit error: %v", err)
	}

	for i := 0; i < 4; i++ {
		if err := scheduler.Submit(Task{
			Name: fmt.Sprintf("queued-%d", i),
			Handler: func(ctx context.Context) (interface{}, error) {
				return "queued", nil
			},
		}); err != nil {
			t.Fatalf("unexpected queued submit error: %v", err)
		}
	}

	cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Submit(Task{
			Name: "after-cancel",
			Handler: func(ctx context.Context) (interface{}, error) {
				return "late", nil
			},
		})
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context canceled, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("submit blocked after context cancellation")
	}
}
