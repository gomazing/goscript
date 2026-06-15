package goscript

import (
	"testing"
	"time"
)

func TestSyncQueueEnqueueAndDrain(t *testing.T) {
	queue := NewSyncQueue()
	received := make(chan []SyncAction, 1)

	queue.Subscribe(func(batch []SyncAction) {
		copyBatch := append([]SyncAction(nil), batch...)
		received <- copyBatch
	})

	queue.Enqueue(SyncAction{ID: "one", Kind: "add"})

	select {
	case batch := <-received:
		if len(batch) != 1 {
			t.Fatalf("expected 1 action in notification, got %d", len(batch))
		}
		if batch[0].ID != "one" {
			t.Fatalf("unexpected action in notification: %+v", batch[0])
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for sync notification")
	}

	drained := queue.Drain()
	if len(drained) != 1 {
		t.Fatalf("expected 1 drained action, got %d", len(drained))
	}
	if drained[0].ID != "one" {
		t.Fatalf("unexpected drained action: %+v", drained[0])
	}

	if snapshot := queue.Snapshot(); len(snapshot) != 0 {
		t.Fatalf("expected empty queue after drain, got %d items", len(snapshot))
	}
}

func TestSyncQueueEnqueueBatch(t *testing.T) {
	queue := NewSyncQueue()
	received := make(chan []SyncAction, 1)

	queue.Subscribe(func(batch []SyncAction) {
		copyBatch := append([]SyncAction(nil), batch...)
		received <- copyBatch
	})

	queue.EnqueueBatch([]SyncAction{
		{ID: "one", Kind: "add"},
		{ID: "two", Kind: "update"},
	})

	select {
	case batch := <-received:
		if len(batch) != 2 {
			t.Fatalf("expected 2 actions in notification, got %d", len(batch))
		}
		if batch[0].ID != "one" || batch[1].ID != "two" {
			t.Fatalf("unexpected batch notification: %+v", batch)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for batch sync notification")
	}

	drained := queue.Drain()
	if len(drained) != 2 {
		t.Fatalf("expected 2 drained actions, got %d", len(drained))
	}
}

func BenchmarkSyncQueueEnqueue(b *testing.B) {
	queue := NewSyncQueue()
	queue.Subscribe(func([]SyncAction) {})

	pending := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(SyncAction{ID: "bench", Kind: "update"})
		pending++
		if pending == 1024 {
			queue.Drain()
			pending = 0
		}
	}
}

func BenchmarkSyncQueueEnqueueBatch(b *testing.B) {
	queue := NewSyncQueue()
	queue.Subscribe(func([]SyncAction) {})

	batch := []SyncAction{
		{ID: "one", Kind: "add"},
		{ID: "two", Kind: "update"},
		{ID: "three", Kind: "update"},
		{ID: "four", Kind: "delete"},
	}

	pending := 0
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.EnqueueBatch(batch)
		pending += len(batch)
		if pending >= 1024 {
			queue.Drain()
			pending = 0
		}
	}
}
