package goscript

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileSnapshot captures the current state of a file for watch mode.
type FileSnapshot struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Hash    string    `json:"hash"`
}

// FileChange describes an added, modified, or removed file.
type FileChange struct {
	Path   string       `json:"path"`
	Kind   string       `json:"kind"`
	Before *FileSnapshot `json:"before,omitempty"`
	After  *FileSnapshot `json:"after,omitempty"`
}

// Watcher polls file roots and emits change batches.
type Watcher struct {
	mu       sync.RWMutex
	roots    []string
	interval time.Duration
	last     map[string]FileSnapshot
}

// NewWatcher creates a new polling watcher.
func NewWatcher(roots ...string) *Watcher {
	cleanRoots := make([]string, 0, len(roots))
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root != "" {
			cleanRoots = append(cleanRoots, root)
		}
	}

	if len(cleanRoots) == 0 {
		cleanRoots = []string{"."}
	}

	return &Watcher{
		roots:    cleanRoots,
		interval: 750 * time.Millisecond,
		last:     map[string]FileSnapshot{},
	}
}

// SetInterval changes the polling interval.
func (w *Watcher) SetInterval(interval time.Duration) {
	if interval <= 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.interval = interval
}

// Snapshot returns the last known file snapshot map.
func (w *Watcher) Snapshot() map[string]FileSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make(map[string]FileSnapshot, len(w.last))
	for path, snapshot := range w.last {
		out[path] = snapshot
	}
	return out
}

// Scan polls all roots and returns the current file state.
func (w *Watcher) Scan() (map[string]FileSnapshot, error) {
	snapshot := make(map[string]FileSnapshot)

	for _, root := range w.roots {
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			fileSnapshot, err := snapshotFile(root)
			if err != nil {
				return nil, err
			}
			snapshot[filepath.Clean(root)] = fileSnapshot
			continue
		}

		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() {
				name := d.Name()
				if path != root && shouldIgnorePath(name) {
					return filepath.SkipDir
				}
				return nil
			}

			if shouldIgnorePath(d.Name()) {
				return nil
			}

			fileSnapshot, err := snapshotFile(path)
			if err != nil {
				return err
			}
			snapshot[filepath.Clean(path)] = fileSnapshot
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return snapshot, nil
}

// Diff compares the last snapshot to a new snapshot and returns changes.
func (w *Watcher) Diff(next map[string]FileSnapshot) []FileChange {
	w.mu.Lock()
	defer w.mu.Unlock()

	changes := diffSnapshots(w.last, next)
	w.last = make(map[string]FileSnapshot, len(next))
	for path, snapshot := range next {
		w.last[path] = snapshot
	}
	return changes
}

// Watch polls the roots until the context is cancelled.
func (w *Watcher) Watch(ctx context.Context) <-chan []FileChange {
	out := make(chan []FileChange)

	go func() {
		defer close(out)

		for {
			snapshot, err := w.Scan()
			if err == nil {
				changes := w.Diff(snapshot)
				if len(changes) > 0 {
					select {
					case out <- changes:
					case <-ctx.Done():
						return
					}
				}
			}

			w.mu.RLock()
			interval := w.interval
			w.mu.RUnlock()

			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()

	return out
}

func snapshotFile(path string) (FileSnapshot, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileSnapshot{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return FileSnapshot{}, err
	}

	sum := sha256.Sum256(data)
	return FileSnapshot{
		Path:    filepath.Clean(path),
		Size:    info.Size(),
		ModTime: info.ModTime().UTC(),
		Hash:    hex.EncodeToString(sum[:]),
	}, nil
}

func diffSnapshots(previous, next map[string]FileSnapshot) []FileChange {
	changes := make([]FileChange, 0)

	for path, after := range next {
		before, existed := previous[path]
		switch {
		case !existed:
			afterCopy := after
			changes = append(changes, FileChange{Path: path, Kind: "added", After: &afterCopy})
		case before.Hash != after.Hash:
			beforeCopy := before
			afterCopy := after
			changes = append(changes, FileChange{Path: path, Kind: "modified", Before: &beforeCopy, After: &afterCopy})
		}
	}

	for path, before := range previous {
		if _, ok := next[path]; !ok {
			beforeCopy := before
			changes = append(changes, FileChange{Path: path, Kind: "removed", Before: &beforeCopy})
		}
	}

	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Kind == changes[j].Kind {
			return changes[i].Path < changes[j].Path
		}
		return changes[i].Kind < changes[j].Kind
	})

	return changes
}

func shouldIgnorePath(name string) bool {
	switch name {
	case "", ".", "..":
		return true
	case ".git", "vendor", "node_modules", "dist", "build", ".idea", ".vscode":
		return true
	default:
		return strings.HasPrefix(name, ".") && name != "."
	}
}

// FormatChange renders a compact human readable summary for watch events.
func FormatChange(change FileChange) string {
	switch change.Kind {
	case "added":
		return fmt.Sprintf("added %s", change.Path)
	case "removed":
		return fmt.Sprintf("removed %s", change.Path)
	default:
		return fmt.Sprintf("modified %s", change.Path)
	}
}
