package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// RotationConfig controls log file rotation behaviour.
type RotationConfig struct {
	Dir       string `yaml:"dir" json:"dir"`
	Filename  string `yaml:"filename" json:"filename"`
	MaxSizeMB int    `yaml:"max_size_mb" json:"maxSizeMB"`
	MaxFiles  int    `yaml:"max_files" json:"maxFiles"`
	MaxAgeDays int   `yaml:"max_age_days" json:"maxAgeDays"`
}

// DefaultRotationConfig returns sensible defaults for file rotation.
func DefaultRotationConfig() RotationConfig {
	return RotationConfig{
		Dir:        "/var/log/cloudos",
		Filename:   "cloudos.log",
		MaxSizeMB:  100,
		MaxFiles:   7,
		MaxAgeDays: 30,
	}
}

// RotatingFileHandler is an io.Writer that rotates log files when they exceed
// a configured size, and purges old rotated files based on age and count.
type RotatingFileHandler struct {
	mu       sync.Mutex
	cfg      RotationConfig
	file     *os.File
	bytesWritten int64
	dir      string
	baseName string
}

// NewRotatingFileHandler creates and opens a RotatingFileHandler.
func NewRotatingFileHandler(cfg RotationConfig) (*RotatingFileHandler, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	h := &RotatingFileHandler{
		cfg:      cfg,
		dir:      cfg.Dir,
		baseName: cfg.Filename,
	}

	if err := h.open(); err != nil {
		return nil, err
	}

	return h, nil
}

// open opens the current log file for appending.
func (h *RotatingFileHandler) open() error {
	path := filepath.Join(h.dir, h.baseName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	info, err := f.Stat()
	if err == nil {
		h.bytesWritten = info.Size()
	}

	h.file = f
	return nil
}

// Write implements io.Writer. It automatically rotates the log file when the
// size limit is exceeded.
func (h *RotatingFileHandler) Write(p []byte) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file == nil {
		return 0, fmt.Errorf("log file is closed")
	}

	maxBytes := int64(h.cfg.MaxSizeMB) * 1024 * 1024
	if h.bytesWritten+int64(len(p)) > maxBytes {
		if err := h.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := h.file.Write(p)
	if err != nil {
		return n, err
	}
	h.bytesWritten += int64(n)
	return n, nil
}

// rotate closes the current file, renames it with a timestamp, opens a new
// file, and purges old archives.
func (h *RotatingFileHandler) rotate() error {
	if h.file != nil {
		if err := h.file.Close(); err != nil {
			return fmt.Errorf("close during rotation: %w", err)
		}
		h.file = nil
	}

	old := filepath.Join(h.dir, h.baseName)
	ts := time.Now().UTC().Format("20060102-150405")
	rotated := filepath.Join(h.dir, fmt.Sprintf("%s.%s", h.baseName, ts))

	if err := os.Rename(old, rotated); err != nil {
		return fmt.Errorf("rename during rotation: %w", err)
	}

	if err := h.open(); err != nil {
		return err
	}

	h.purge()
	return nil
}

// purge removes old rotated files beyond MaxFiles and MaxAgeDays limits.
func (h *RotatingFileHandler) purge() {
	entries, err := os.ReadDir(h.dir)
	if err != nil {
		return
	}

	type rotatedFile struct {
		path    string
		modTime time.Time
	}

	var files []rotatedFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Match files named baseName.TIMESTAMP
		if len(e.Name()) > len(h.baseName) && e.Name()[:len(h.baseName)] == h.baseName && e.Name()[len(h.baseName)] == '.' {
			info, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, rotatedFile{
				path:    filepath.Join(h.dir, e.Name()),
				modTime: info.ModTime(),
			})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove by age.
	cutoff := time.Now().AddDate(0, 0, -h.cfg.MaxAgeDays)
	for _, f := range files {
		if f.modTime.Before(cutoff) {
			os.Remove(f.path)
		}
	}

	// Remove by count (re-filter after age removal).
	if h.cfg.MaxFiles > 0 && len(files) > h.cfg.MaxFiles {
		for i := 0; i < len(files)-h.cfg.MaxFiles; i++ {
			os.Remove(files[i].path)
		}
	}
}

// Close closes the current log file.
func (h *RotatingFileHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file != nil {
		return h.file.Close()
	}
	return nil
}

// Compile-time interface check.
var _ io.Writer = (*RotatingFileHandler)(nil)
