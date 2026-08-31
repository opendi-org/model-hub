package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration sourced from environment variables.
// Every field maps directly to a variable defined in docker-compose.
type Config struct {
	// Database
	DBHostname    string
	DBPort        int
	DBName        string
	DBUsername    string
	DBPassword    string
	DBSSLMode     string
	DBSSLRootCert string

	// Server
	Address string // MODEL_HUB_ADDRESS
	Port    int    // MODEL_HUB_PORT

	// Auth
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Runtime
	DevMode bool
}

// DSN returns a PostgreSQL connection string for GORM.
func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.DBHostname, c.DBPort, c.DBUsername, c.DBPassword, c.DBName, c.DBSSLMode,
	)
	if c.DBSSLRootCert != "" {
		dsn += fmt.Sprintf(
			" sslrootcert=%s", c.DBSSLRootCert,
		)
	}
	return dsn
}

// ListenAddr returns the host:port string for the HTTP server.
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// LoadConfig reads all required environment variables and returns a Config.
// Returns an error listing every missing or invalid variable so the operator
// sees all problems at once rather than one at a time.
func LoadConfig() (*Config, error) {
	var missing []string

	optional := func(key, defaultValue string) string {
		v := os.Getenv(key)
		if v == "" {
			return defaultValue
		}
		return v
	}

	require := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	requireInt := func(key string, fallback int) int {
		v := os.Getenv(key)
		if v == "" {
			return fallback
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			missing = append(missing, key+" (must be integer)")
			return fallback
		}
		return n
	}
	devMode := os.Getenv("DEV_MODE") == "true"

	cfg := &Config{
		DBHostname:    require("DB_HOSTNAME"),
		DBPort:        requireInt("DB_PORT", 5432),
		DBName:        require("DB_NAME"),
		DBUsername:    require("DB_USERNAME"),
		DBPassword:    require("DB_PASSWORD"),
		DBSSLMode:     optional("DB_SSL_MODE", "disable"),
		DBSSLRootCert: optional("DB_SSL_ROOT_CERT", ""),
		Address:       optional("MODEL_HUB_ADDRESS", "0.0.0.0"),
		Port:          requireInt("MODEL_HUB_PORT", 8080),

		JWTSecret:          require("JWT_SECRET"),
		GoogleClientID:     require("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: require("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  require("GOOGLE_REDIRECT_URL"),

		DevMode: devMode,
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing or invalid environment variables: %v", missing)
	}

	return cfg, nil
}
