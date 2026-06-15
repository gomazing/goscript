package goscript

import (
	"sync"
	"time"
)

// SyncAction describes a queued offline-first update.
type SyncAction struct {
	ID        string                 `json:"id"`
	Scope     string                 `json:"scope"`
	Kind      string                 `json:"kind"`
	Payload   interface{}            `json:"payload"`
	CreatedAt time.Time              `json:"createdAt"`
	Meta      map[string]string      `json:"meta,omitempty"`
}

// SyncQueue stores actions until they can be flushed.
type SyncQueue struct {
	mu        sync.RWMutex
	items     []SyncAction
	listeners []func([]SyncAction)
}

// NewSyncQueue creates an empty queue.
func NewSyncQueue() *SyncQueue {
	return &SyncQueue{
		items: make([]SyncAction, 0),
	}
}

// Enqueue adds a new sync action.
func (q *SyncQueue) Enqueue(action SyncAction) {
	q.EnqueueBatch([]SyncAction{action})
}

// EnqueueBatch adds multiple sync actions with one lock pass.
func (q *SyncQueue) EnqueueBatch(actions []SyncAction) {
	if len(actions) == 0 {
		return
	}

	stamp := time.Now().UTC()
	normalized := make([]SyncAction, 0, len(actions))
	listeners := make([]func([]SyncAction), 0)

	q.mu.Lock()
	for _, action := range actions {
		if action.CreatedAt.IsZero() {
			action.CreatedAt = stamp
		}
		q.items = append(q.items, action)
		normalized = append(normalized, action)
	}
	listeners = append(listeners, q.listeners...)
	q.mu.Unlock()

	if len(listeners) == 0 {
		return
	}

	batch := append([]SyncAction(nil), normalized...)
	for _, listener := range listeners {
		if listener != nil {
			go listener(batch)
		}
	}
}

// Drain returns all queued actions and clears the queue.
func (q *SyncQueue) Drain() []SyncAction {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := q.items
	q.items = nil
	return out
}

// Snapshot returns a copy of queued actions without clearing them.
func (q *SyncQueue) Snapshot() []SyncAction {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return append([]SyncAction(nil), q.items...)
}

// Subscribe registers a queue listener that receives newly enqueued batches.
func (q *SyncQueue) Subscribe(listener func([]SyncAction)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.listeners = append(q.listeners, listener)
}
