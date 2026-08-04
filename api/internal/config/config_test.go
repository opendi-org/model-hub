package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSN_UsesConfiguredSSLMode(t *testing.T) {
	cfg := &Config{
		DBHostname: "db",
		DBPort:     5432,
		DBUsername: "postgres",
		DBPassword: "secret",
		DBName:     "model_hub",
		DBSSLMode:  "verify-full",
	}

	dsn := cfg.DSN()

	assert.Contains(t, dsn, "sslmode=verify-full")
}

func TestLoadConfig_DBSSLMode(t *testing.T) {
	// required vars LoadConfig() needs to succeed at all
	t.Setenv("DB_HOSTNAME", "db")
	t.Setenv("DB_NAME", "model_hub")
	t.Setenv("DB_USERNAME", "postgres")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("GOOGLE_CLIENT_ID", "id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "secret")
	t.Setenv("GOOGLE_REDIRECT_URL", "http://localhost/callback")

	t.Run("defaults to disable when unset", func(t *testing.T) {
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "disable", cfg.DBSSLMode)
	})

	t.Run("honors DB_SSL_MODE override", func(t *testing.T) {
		t.Setenv("DB_SSL_MODE", "verify-full")
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "verify-full", cfg.DBSSLMode)
	})
}
