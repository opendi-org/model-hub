package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	} else {
		if cfg.DBSSLMode == "disable" {
			log.Println("WARNING: DB_SSL_MODE is 'disable' in production mode. This is insecure! Recommend enabling SSL for production.")
		}
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

// SeedDemoData populates the database with realistic demo repositories when DEV_MODE is true.
// It creates an "opendi" user with 12 public model repositories, each with a tag pointing to an
// example CDM model from cdm-json-schema/examples.
// Safe to call multiple times — only creates data if it doesn't already exist.
func SeedDemoData(db *gorm.DB) error {
	log.Println("seeding demo data...")

	// Check if "opendi" user already exists
	var existingUser hub.User
	if err := db.Where("username = ?", "opendi").First(&existingUser).Error; err == nil {
		// User already exists, skip seeding
		log.Println("demo data already seeded, skipping")
		return nil
	}

	// Create the "opendi" user
	user := hub.User{
		Username:    "opendi",
		Email:       "demo@opendi.org",
		DisplayName: "OpenDI Demo",
		Bio:         "Demo account with example decision intelligence models",
	}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("creating opendi user: %w", err)
	}
	log.Printf("created user opendi (id=%d)", user.ID)

	// Define demo repositories: (slug, description, exampleFile, tagName)
	demoRepos := []struct {
		slug        string
		description string
		exampleFile string
		tagName     string
	}{
		{
			slug:        "portfolio-optimizer",
			description: "Analyze investment portfolios across asset classes. Visualize allocation across multiple investment categories to optimize risk and return.",
			exampleFile: "Bar_Graph_Demo_2D_Array.json",
			tagName:     "latest",
		},
		{
			slug:        "supply-chain-analyzer",
			description: "Evaluate supply chain costs and efficiency across different regions and suppliers. Compare performance metrics dynamically.",
			exampleFile: "Bar_Graph_Demo_Dynamic_Categories.json",
			tagName:     "latest",
		},
		{
			slug:        "risk-assessment",
			description: "Quantify organizational risks across operational, financial, and strategic domains. Adjust risk scores using interactive sliders.",
			exampleFile: "Range_Demo_Basic_Adder.json",
			tagName:     "latest",
		},
		{
			slug:        "pricing-strategy-simulator",
			description: "Model pricing strategies with dynamic market conditions. Adjust price ranges to evaluate revenue and margin outcomes.",
			exampleFile: "Range_Demo_Dynamic_Range.json",
			tagName:     "latest",
		},
		{
			slug:        "workforce-scheduler",
			description: "Schedule workforce across departments and shifts. Select team members for optimal coverage and utilization.",
			exampleFile: "Selector_Demo_Basic.json",
			tagName:     "latest",
		},
		{
			slug:        "dynamic-allocator",
			description: "Dynamically allocate resources based on real-time demand signals. Interactive selector for flexible resource distribution.",
			exampleFile: "Selector_Demo_Dynamic.json",
			tagName:     "latest",
		},
		{
			slug:        "budget-planner",
			description: "Plan organizational budgets through multi-step allocation decisions. Build departmental budgets incrementally with clear visualization.",
			exampleFile: "Range_Demo_Multistep_Adder.json",
			tagName:     "latest",
		},
		{
			slug:        "satisfaction-predictor",
			description: "Predict customer satisfaction based on service attributes and operational choices. Combine text inputs with boolean flags for nuanced analysis.",
			exampleFile: "Text_and_Bool_Demo_Word_Combination.json",
			tagName:     "latest",
		},
		{
			slug:        "personnel-development",
			description: "Track and plan personnel development programs. Use text descriptions and boolean indicators for comprehensive tracking.",
			exampleFile: "Text_and_Bool_Demo_Word_Combination_Multi.json",
			tagName:     "latest",
		},
		{
			slug:        "cafe-operations",
			description: "Optimize cafe operations with interactive scenario modeling. Analyze workflow and resource utilization in real-time.",
			exampleFile: "coffee.json",
			tagName:     "v1.0",
		},
		{
			slug:        "demand-forecaster",
			description: "Generate deterministic demand forecasts for inventory planning. Non-interactive model for batch processing and analysis.",
			exampleFile: "coffee_noninteractive.json",
			tagName:     "v1.0",
		},
		{
			slug:        "market-analyzer",
			description: "Analyze market conditions and competitive positioning. Interactive model for scenario analysis and decision support.",
			exampleFile: "coffee.json",
			tagName:     "v1.0",
		},
	}

	// Load example files and create repositories with tags
	examplesDir := "cdm-json-schema/examples"
	for _, repo := range demoRepos {
		// Read the example CDM JSON file
		examplePath := filepath.Join(examplesDir, repo.exampleFile)
		raw, err := os.ReadFile(examplePath)
		if err != nil {
			return fmt.Errorf("reading example file %s: %w", repo.exampleFile, err)
		}

		// Save the CDM and get its root UUID
		modelUUID, err := SaveCDM(db, raw)
		if err != nil {
			return fmt.Errorf("saving CDM from %s: %w", repo.exampleFile, err)
		}

		// Create the repository
		repository := hub.Repository{
			OwnerID:     user.ID,
			Slug:        repo.slug,
			Description: repo.description,
			Visibility:  "public",
		}
		if err := db.Create(&repository).Error; err != nil {
			return fmt.Errorf("creating repository %s: %w", repo.slug, err)
		}

		// Create a tag pointing to the model
		tag := hub.CDMTag{
			RepoID:      repository.ID,
			Name:        repo.tagName,
			ModelUUID:   modelUUID,
			SizeBytes:   int64(len(raw)),
			CreatedByID: user.ID,
		}
		if err := db.Create(&tag).Error; err != nil {
			return fmt.Errorf("creating tag for repo %s: %w", repo.slug, err)
		}

		log.Printf("created repository: opendi/%s with tag %s pointing to %s", repo.slug, repo.tagName, repo.exampleFile)
	}

	log.Println("demo data seeding complete")
	return nil
}
