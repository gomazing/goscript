package goscript

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOptionAndResult(t *testing.T) {
	opt := Some("hello")
	if got := opt.UnwrapOr("fallback"); got != "hello" {
		t.Fatalf("expected hello, got %v", got)
	}

	res := Ok(42)
	if !res.IsOk() {
		t.Fatalf("expected result to be ok")
	}
	if got := res.UnwrapOr(0); got != 42 {
		t.Fatalf("expected 42, got %v", got)
	}

	if Match("alpha",
		MatchCase{Equals: "beta", Then: func(v interface{}) interface{} { return "nope" }},
		MatchCase{Kind: 0, Then: func(v interface{}) interface{} { return "any" }},
	) != nil {
		t.Fatalf("unexpected match result")
	}
}

func TestScheduler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := NewScheduler(2)
	scheduler.Start(ctx)
	defer scheduler.Stop()

	if err := scheduler.Submit(Task{
		Name: "echo",
		Handler: func(ctx context.Context) (interface{}, error) {
			return "ok", nil
		},
	}); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	select {
	case result := <-scheduler.Results():
		if result.Err != nil || result.Value != "ok" {
			t.Fatalf("unexpected result: %+v", result)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for task result")
	}
}

func TestRealtimeHub(t *testing.T) {
	hub := NewRealtimeHub(2)
	sub := hub.Subscribe("system", "listener-1")
	defer hub.Unsubscribe("system", "listener-1")

	hub.Ping("system", "unit-test")

	select {
	case event := <-sub:
		if event.Kind != "ping" {
			t.Fatalf("expected ping event, got %q", event.Kind)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for ping event")
	}

	if events := hub.Poll("system", 10); len(events) == 0 {
		t.Fatalf("expected history to contain events")
	}
}

func TestRealtimeHubUnsubscribeThenPublish(t *testing.T) {
	hub := NewRealtimeHub(1)
	sub := hub.Subscribe("system", "listener-1")
	hub.Unsubscribe("system", "listener-1")

	select {
	case <-sub:
	default:
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("publish should not panic after unsubscribe: %v", r)
		}
	}()

	hub.Publish("system", "ping", map[string]interface{}{"alive": true}, "unit-test")
}

func TestRealtimeHubPublishBatch(t *testing.T) {
	hub := NewRealtimeHub(4)
	sub := hub.Subscribe("sheet", "listener-1")
	defer hub.Unsubscribe("sheet", "listener-1")

	events := hub.PublishBatch("sheet", "unit-test", []StreamEvent{
		{Kind: "cell", Payload: map[string]string{"cell": "A1"}},
		{Kind: "cell", Payload: map[string]string{"cell": "B1"}},
	})
	if len(events) != 2 {
		t.Fatalf("expected two normalized events, got %d", len(events))
	}

	for i := 0; i < 2; i++ {
		select {
		case event := <-sub:
			if event.Topic != "sheet" || event.Source != "unit-test" {
				t.Fatalf("unexpected batch event: %+v", event)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for batch event %d", i)
		}
	}
}

func BenchmarkRealtimeHubPublishBatch(b *testing.B) {
	hub := NewRealtimeHub(64)
	sub := hub.Subscribe("sheet", "listener-1")
	defer hub.Unsubscribe("sheet", "listener-1")

	batch := make([]StreamEvent, 0, 32)
	for i := 0; i < 32; i++ {
		batch = append(batch, StreamEvent{
			Kind:    "cell",
			Payload: map[string]string{"cell": "A1"},
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		published := hub.PublishBatch("sheet", "bench", batch)
		if len(published) != len(batch) {
			b.Fatalf("expected %d published events, got %d", len(batch), len(published))
		}
		for j := 0; j < len(batch); j++ {
			select {
			case <-sub:
			default:
				b.Fatalf("expected subscriber event %d", j)
			}
		}
	}
}

func TestInferenceRouter(t *testing.T) {
	router := NewInferenceRouter(
		fakeProvider{label: "local"},
		nil,
		fakeProvider{label: "fallback"},
	)

	response, err := router.Infer(context.Background(), InferenceRequest{Model: "tiny"})
	if err != nil {
		t.Fatalf("unexpected inference error: %v", err)
	}
	if response.Provider != "local" {
		t.Fatalf("expected local provider, got %q", response.Provider)
	}
}

type fakeProvider struct {
	label string
}

func (f fakeProvider) Infer(ctx context.Context, request InferenceRequest) (InferenceResponse, error) {
	if request.Model == "fail" {
		return InferenceResponse{}, errors.New("forced failure")
	}

	return InferenceResponse{
		Model:    request.Model,
		Output:   request.Input,
		Provider: f.label,
	}, nil
}
