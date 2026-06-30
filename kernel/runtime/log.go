package runtime

import (
	"bytes"
	"context"
	"sync"
	"time"
)

// ── Log Entry ──────────────────────────────────────────────────────────────

// LogEntry represents a single log line from an application instance.
type LogEntry struct {
	// Timestamp is when the log line was captured.
	Timestamp time.Time `json:"timestamp"`

	// Source identifies where the log came from ("stdout", "stderr").
	Source string `json:"source"`

	// Line is the log message content.
	Line string `json:"line"`

	// InstanceID is the ID of the application instance that produced this log.
	InstanceID string `json:"instanceId"`

	// AppID is the CloudOS Application ID.
	AppID string `json:"appId"`
}

// ── Log Store ──────────────────────────────────────────────────────────────

// LogStore is a ring buffer for log entries. It is safe for concurrent access.
type LogStore struct {
	mu      sync.Mutex
	cond    *sync.Cond
	entries []LogEntry
	offset  int
	count   int
	cap     int

	// pending tracks whether any followers are waiting.
	followers map[chan LogEntry]struct{}
}

// NewLogStore creates a ring buffer with the given capacity.
func NewLogStore(capacity int) *LogStore {
	if capacity <= 0 {
		capacity = 1000
	}
	s := &LogStore{
		entries:   make([]LogEntry, capacity),
		cap:       capacity,
		followers: make(map[chan LogEntry]struct{}),
	}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// Write adds a log entry to the store.
func (s *LogStore) Write(entry LogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[s.offset] = entry
	s.offset = (s.offset + 1) % s.cap
	if s.count < s.cap {
		s.count++
	}

	// Notify followers.
	for ch := range s.followers {
		select {
		case ch <- entry:
		default:
			// Drop if follower is slow.
		}
	}

	s.cond.Broadcast()
}

// Read returns up to n log entries, oldest first.
// If n <= 0, returns all available entries.
func (s *LogStore) Read(n int) []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.count == 0 {
		return nil
	}

	limit := s.count
	if n > 0 && n < limit {
		limit = n
	}

	result := make([]LogEntry, limit)
	if s.count < s.cap {
		// Buffer hasn't wrapped yet.
		start := 0
		if limit < s.count {
			start = s.count - limit
		}
		copy(result, s.entries[start:s.count])
	} else {
		// Buffer has wrapped; start from offset.
		avail := s.cap
		start := 0
		if limit < avail {
			start = avail - limit
		}
		idx := 0
		for i := (s.offset + start) % s.cap; idx < limit; i = (i + 1) % s.cap {
			result[idx] = s.entries[i]
			idx++
		}
	}
	return result
}

// Clear clears all entries from the store.
func (s *LogStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset = 0
	s.count = 0
	s.entries = make([]LogEntry, s.cap)
}

// Follow returns a channel that receives new log entries as they are written.
// The channel is closed when the context is cancelled.
func (s *LogStore) Follow(ctx context.Context) <-chan LogEntry {
	ch := make(chan LogEntry, 64)

	s.mu.Lock()
	s.followers[ch] = struct{}{}
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.followers, ch)
		close(ch)
		s.mu.Unlock()
	}()

	return ch
}

// Len returns the number of entries currently stored.
func (s *LogStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// ── Log Manager ────────────────────────────────────────────────────────────

// LogManager is the central log aggregator for CloudOS.
// Every runtime writes to the LogManager. It provides:
//   - Ring-buffer storage for recent logs
//   - Follow/stream for live tailing
//   - Per-application and per-instance filtering
//
// Architecture:
//
//	Runtime → LogManager.Write()
//	              │
//	              ├── Per-app LogStore (ring buffer)
//	              ├── Follow channels for SSE streaming
//	              └── REST API reads via Read()
type LogManager struct {
	mu    sync.RWMutex
	stores map[string]*LogStore // keyed by appID
	cap   int
}

// NewLogManager creates a new LogManager with the given per-app capacity.
func NewLogManager(capacity int) *LogManager {
	if capacity <= 0 {
		capacity = 1000
	}
	return &LogManager{
		stores: make(map[string]*LogStore),
		cap:    capacity,
	}
}

// Write writes a log entry for an application.
// It creates the store for the app if it doesn't exist yet.
func (lm *LogManager) Write(entry LogEntry) {
	store := lm.getOrCreateStore(entry.AppID)
	store.Write(entry)
}

// Read returns recent log entries for an application.
func (lm *LogManager) Read(appID string, limit int) []LogEntry {
	store := lm.getStore(appID)
	if store == nil {
		return nil
	}
	return store.Read(limit)
}

// Follow returns a channel that streams new log entries for an app.
func (lm *LogManager) Follow(ctx context.Context, appID string) <-chan LogEntry {
	store := lm.getOrCreateStore(appID)
	return store.Follow(ctx)
}

// Clear clears all logs for an application.
func (lm *LogManager) Clear(appID string) {
	store := lm.getStore(appID)
	if store != nil {
		store.Clear()
	}
}

// DeleteStore removes the log store for an application.
func (lm *LogManager) DeleteStore(appID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	delete(lm.stores, appID)
}

func (lm *LogManager) getStore(appID string) *LogStore {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.stores[appID]
}

func (lm *LogManager) getOrCreateStore(appID string) *LogStore {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	store, ok := lm.stores[appID]
	if !ok {
		store = NewLogStore(lm.cap)
		lm.stores[appID] = store
	}
	return store
}

// ── LogLineWriter ──────────────────────────────────────────────────────────

// LogLineWriter implements io.Writer for capturing output from a process
// and writing structured LogEntry records to the LogManager.
type LogLineWriter struct {
	Manager    *LogManager
	AppID      string
	InstanceID string
	Source     string // "stdout" or "stderr"
	buffer     bytes.Buffer
}

func (w *LogLineWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.buffer.Write(p)

	for {
		data := w.buffer.String()
		idx := bytes.IndexByte([]byte(data), '\n')
		if idx < 0 {
			break
		}

		line := data[:idx]
		w.buffer.Next(idx + 1)

		if line != "" {
			w.Manager.Write(LogEntry{
				Timestamp:  time.Now(),
				Source:     w.Source,
				Line:       line,
				InstanceID: w.InstanceID,
				AppID:      w.AppID,
			})
		}
	}
	return n, nil
}
