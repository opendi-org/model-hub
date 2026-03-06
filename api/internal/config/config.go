package config

import (
	"os"
	"strconv"
)

// Config holds application configuration from environment (e.g. compose.yaml).
type Config struct {
	// Database (Postgres)
	DBHostname string
	DBPort     int
	DBName     string
	DBUsername string
	DBPassword string

	// Server
	ModelHubAddress string
	ModelHubPort    string

	// Auth
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Dev
	DevMode bool
}

// Load reads configuration from the environment.
func Load() *Config {
	port := 5432
	if p := os.Getenv("DB_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	devMode := os.Getenv("DEV_MODE") == "true" || os.Getenv("DEV_MODE") == "1"
	return &Config{
		DBHostname:         getEnv("DB_HOSTNAME", "localhost"),
		DBPort:             port,
		DBName:             getEnv("DB_NAME", "modelhub"),
		DBUsername:         getEnv("DB_USERNAME", "postgres"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		ModelHubAddress:    getEnv("MODEL_HUB_ADDRESS", "localhost"),
		ModelHubPort:       getEnv("MODEL_HUB_PORT", "8080"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		DevMode:            devMode,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// DSN returns a PostgreSQL connection string suitable for gorm postgres.Open.
func (c *Config) DSN() string {
	// sslmode=disable is typical for local/docker Postgres; use require for production.
	sslmode := "disable"
	if !c.DevMode {
		sslmode = "require"
	}
	return "host=" + c.DBHostname +
		" port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUsername +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + sslmode
}
