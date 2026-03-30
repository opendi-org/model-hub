package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/models/hub"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Warn
	if cfg.DevMode {
		logLevel = logger.Info // log all SQL in dev
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// Migrate runs AutoMigrate for all tables.
// Safe to call on every startup — only adds columns/tables, never drops.
func Migrate(db *gorm.DB) error {
	log.Println("running database migrations...")
	return db.AutoMigrate(

		// Immutable CDM content tables.
		&CDMModel{},
		&CDMRunnableModel{},
		&CDMDiagram{},
		&CDMEvaluatableAsset{},
		&CDMIOValue{},
		&CDMControl{},

		// Explicit join tables: CDMModel <-> children (managed manually, no GORM associations).
		&CDMModelRunnableModel{},
		&CDMModelDiagram{},
		&CDMModelEvaluatableAsset{},
		&CDMModelIOValue{},
		&CDMModelControl{},

		// Mutable hub-layer tables.
		&hub.User{},
		&hub.OAuthIdentity{},
		&hub.CLISession{},
		&hub.Repository{},
		&hub.Collaborator{},
		&hub.CDMTag{},
	)
}
