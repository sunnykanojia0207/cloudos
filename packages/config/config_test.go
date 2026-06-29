package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "info", cfg.Kernel.LogLevel)
	assert.Equal(t, 8080, cfg.API.Port)
	assert.Equal(t, "24h", cfg.Auth.TokenTTL)
}

func TestInterpolateEnv(t *testing.T) {
	t.Run("simple variable", func(t *testing.T) {
		os.Setenv("CLOUDOS_TEST_VAL", "resolved")
		defer os.Unsetenv("CLOUDOS_TEST_VAL")

		result := string(interpolateEnv([]byte("${CLOUDOS_TEST_VAL}")))
		assert.Equal(t, "resolved", result)
	})

	t.Run("variable with default", func(t *testing.T) {
		result := string(interpolateEnv([]byte("${UNSET_VAR:-default}")))
		assert.Equal(t, "default", result)
	})

	t.Run("variable with default, env set", func(t *testing.T) {
		os.Setenv("CLOUDOS_TEST_PORT", "9090")
		defer os.Unsetenv("CLOUDOS_TEST_PORT")

		result := string(interpolateEnv([]byte("${CLOUDOS_TEST_PORT:-8080}")))
		assert.Equal(t, "9090", result)
	})

	t.Run("no interpolation", func(t *testing.T) {
		result := string(interpolateEnv([]byte("plain text")))
		assert.Equal(t, "plain text", result)
	})

	t.Run("multiple variables", func(t *testing.T) {
		os.Setenv("H", "a")
		os.Setenv("P", "b")
		defer os.Unsetenv("H")
		defer os.Unsetenv("P")

		result := string(interpolateEnv([]byte("${H}:${P}")))
		assert.Equal(t, "a:b", result)
	})

	t.Run("unset variable without default is preserved", func(t *testing.T) {
		result := string(interpolateEnv([]byte("${NOT_SET}")))
		assert.Equal(t, "${NOT_SET}", result)
	})
}

func TestYAMLProviderLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cloudos.yaml")

	content := `
kernel:
  log_level: debug
  data_dir: /tmp/cloudos
api:
  host: 127.0.0.1
  port: 9090
auth:
  jwt_secret: test-secret
  token_ttl: 1h
logging:
  level: debug
  format: text
  output: stdout
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	p := NewYAMLProvider()
	cfg, err := p.Load(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "debug", cfg.Kernel.LogLevel)
	assert.Equal(t, "/tmp/cloudos", cfg.Kernel.DataDir)
	assert.Equal(t, "127.0.0.1", cfg.API.Host)
	assert.Equal(t, 9090, cfg.API.Port)
	assert.Equal(t, "test-secret", cfg.Auth.JWTSecret)
	assert.Equal(t, "1h", cfg.Auth.TokenTTL)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

func TestYAMLProviderLoadWithEnv(t *testing.T) {
	os.Setenv("CLOUDOS_TEST_PORT", "7070")
	defer os.Unsetenv("CLOUDOS_TEST_PORT")

	dir := t.TempDir()
	path := filepath.Join(dir, "cloudos.yaml")

	content := `
api:
  host: ${CLOUDOS_TEST_HOST:-0.0.0.0}
  port: ${CLOUDOS_TEST_PORT:-8080}
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	p := NewYAMLProvider()
	cfg, err := p.Load(path)
	require.NoError(t, err)

	assert.Equal(t, 7070, cfg.API.Port)
	assert.Equal(t, "0.0.0.0", cfg.API.Host)
}

func TestYAMLProviderLoadFileNotFound(t *testing.T) {
	p := NewYAMLProvider()
	_, err := p.Load("/nonexistent/path.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read config")
}

func TestYAMLProviderGet(t *testing.T) {
	p := NewYAMLProvider()
	cfg := p.Get()
	assert.Nil(t, cfg, "Get should return nil before Load")
}

func TestYAMLProviderClose(t *testing.T) {
	p := NewYAMLProvider()
	err := p.Close()
	assert.NoError(t, err)
}
