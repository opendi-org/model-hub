package database

import "gorm.io/gorm"

// RunMigrations creates or updates CDM tables only (no hub tables).
// Reversible migrations (Down) can be added later if needed.
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		// Core CDM tables (one row per unique content-hashed entity).
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
	)
}
