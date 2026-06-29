// Package security provides the security context manager for the CloudOS kernel.
// It manages authentication and authorisation primitives used by the kernel to
// validate operations against configured policies.
package security

import (
	"context"
	"sync"
	"time"
)

// Principal represents an authenticated entity (user, service, system).
type Principal struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "user", "service", "system"
	Roles     []string  `json:"roles"`
	AuthdAt   time.Time `json:"authenticatedAt"`
}

// Context carries the security context for a request.
type Context struct {
	Principal Principal
	Valid     bool
}

// Manager provides authentication and authorisation primitives.
type Manager struct {
	mu        sync.RWMutex
	apiTokens map[string]string // token -> principal ID
}

// NewManager creates a new security manager.
func NewManager() *Manager {
	return &Manager{
		apiTokens: make(map[string]string),
	}
}

// AuthenticateToken validates an API token and returns a security context.
func (m *Manager) AuthenticateToken(ctx context.Context, token string) (Context, error) {
	m.mu.RLock()
	id, ok := m.apiTokens[token]
	m.mu.RUnlock()

	if !ok {
		return Context{Valid: false}, nil
	}

	return Context{
		Principal: Principal{
			ID:      id,
			Type:    "service",
			Roles:   []string{"admin"},
			AuthdAt: time.Now(),
		},
		Valid: true,
	}, nil
}

// RegisterToken registers an API token for the given principal ID.
func (m *Manager) RegisterToken(token, principalID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.apiTokens[token] = principalID
}

// RevokeToken removes an API token.
func (m *Manager) RevokeToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.apiTokens, token)
}

// HasRole checks whether a security context includes the given role.
func HasRole(ctx Context, role string) bool {
	if !ctx.Valid {
		return false
	}
	for _, r := range ctx.Principal.Roles {
		if r == role {
			return true
		}
	}
	return false
}
