// Package storage provides the built-in storage.local provider that stores
// objects on the local filesystem.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudos/cloudos/capabilities"
	"github.com/cloudos/cloudos/providers"
)

const (
	providerName    = "storage.local"
	providerVersion = "0.1.0"
	providerDesc    = "Built-in local filesystem storage provider"
)

// LocalProvider stores objects on the local filesystem.
type LocalProvider struct {
	mu       sync.Mutex
	state    providers.State
	basePath string
}

// NewLocalProvider creates a new local storage provider.
func NewLocalProvider(basePath string) *LocalProvider {
	return &LocalProvider{
		state:    providers.StateDiscovered,
		basePath: basePath,
	}
}

// Info returns provider metadata.
func (p *LocalProvider) Info() providers.Info {
	return providers.Info{
		Name:        providerName,
		Version:     providerVersion,
		Description: providerDesc,
		Capability:  "storage",
	}
}

// Init configures the provider.
func (p *LocalProvider) Init(ctx context.Context, config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = providers.StateInit
	return nil
}

// Start transitions the provider to the ready state.
func (p *LocalProvider) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := os.MkdirAll(p.basePath, 0755); err != nil {
		return fmt.Errorf("create storage dir: %w", err)
	}

	p.state = providers.StateReady
	return nil
}

// Stop shuts down the provider.
func (p *LocalProvider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = providers.StateStopped
	return nil
}

// State returns the current provider state.
func (p *LocalProvider) State() providers.State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// Capability returns the storage capability.
func (p *LocalProvider) Capability() capabilities.Capability {
	return p
}

// --- Capability interface implementation -----------------------------------

func (p *LocalProvider) ID() capabilities.ID             { return "storage" }
func (p *LocalProvider) Version() capabilities.Version    { return capabilities.Version{Major: 1, Minor: 0, Patch: 0} }

func (p *LocalProvider) Validate(ctx context.Context) error {
	if p.state != providers.StateReady {
		return fmt.Errorf("provider not ready")
	}
	return nil
}

func (p *LocalProvider) Put(ctx context.Context, bucket string, key string, data []byte, contentType string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dir := filepath.Join(p.basePath, bucket)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create bucket dir: %w", err)
	}

	path := filepath.Join(dir, key)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write object: %w", err)
	}
	return nil
}

func (p *LocalProvider) Get(ctx context.Context, bucket string, key string) ([]byte, error) {
	path := filepath.Join(p.basePath, bucket, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}
	return data, nil
}

func (p *LocalProvider) Delete(ctx context.Context, bucket string, key string) error {
	path := filepath.Join(p.basePath, bucket, key)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (p *LocalProvider) List(ctx context.Context, bucket string, prefix string) ([]*capabilities.Object, error) {
	dir := filepath.Join(p.basePath, bucket)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list bucket: %w", err)
	}

	var objects []*capabilities.Object
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		objects = append(objects, &capabilities.Object{
			Key:          e.Name(),
			Size:         info.Size(),
			ContentType:  detectContentType(e.Name()),
			LastModified: info.ModTime().Unix(),
		})
	}

	if objects == nil {
		objects = []*capabilities.Object{}
	}
	return objects, nil
}

func (p *LocalProvider) CreateBucket(ctx context.Context, bucket string) error {
	dir := filepath.Join(p.basePath, bucket)
	return os.MkdirAll(dir, 0755)
}

func (p *LocalProvider) DeleteBucket(ctx context.Context, bucket string) error {
	dir := filepath.Join(p.basePath, bucket)
	return os.RemoveAll(dir)
}

// detectContentType returns a basic content type based on file extension.
func detectContentType(name string) string {
	ext := filepath.Ext(name)
	switch ext {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html", ".htm":
		return "text/html"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
