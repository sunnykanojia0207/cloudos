package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndAuthenticateToken(t *testing.T) {
	m := NewManager()
	m.RegisterToken("tok-123", "svc-1")

	ctx := context.Background()
	c, err := m.AuthenticateToken(ctx, "tok-123")
	require.NoError(t, err)
	assert.True(t, c.Valid)
	assert.Equal(t, "svc-1", c.Principal.ID)
}

func TestInvalidToken(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	c, err := m.AuthenticateToken(ctx, "invalid")
	require.NoError(t, err)
	assert.False(t, c.Valid)
}

func TestRevokeToken(t *testing.T) {
	m := NewManager()
	m.RegisterToken("tok-abc", "svc-2")
	m.RevokeToken("tok-abc")

	ctx := context.Background()
	c, err := m.AuthenticateToken(ctx, "tok-abc")
	require.NoError(t, err)
	assert.False(t, c.Valid)
}

func TestHasRole(t *testing.T) {
	m := NewManager()
	m.RegisterToken("tok-admin", "admin-svc")

	ctx := context.Background()

	adminCtx, _ := m.AuthenticateToken(ctx, "tok-admin")
	assert.True(t, HasRole(adminCtx, "admin"))
	assert.False(t, HasRole(adminCtx, "nonexistent"))

	invalidCtx := Context{Valid: false}
	assert.False(t, HasRole(invalidCtx, "admin"))
}
