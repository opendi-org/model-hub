package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open opens a PostgreSQL database connection. DSN format: "host=... user=... password=... dbname=... sslmode=..."
func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
