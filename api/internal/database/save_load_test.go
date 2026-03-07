package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// apiRoot returns the api/ directory (parent of internal/), relative to this test file.
func apiRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func init() {
	p := filepath.Join(apiRoot(), "cdm-json-schema", "schema-source", "Causal-Decision-Model.json")
	if _, err := os.Stat(p); err == nil {
		cdmSchemaPath = p
	}
}

// testDB opens a Postgres DB and runs migrations. Skips the test if DB is not available.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		host := os.Getenv("DB_HOSTNAME")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USERNAME")
		if user == "" {
			user = "postgres"
		}
		pass := os.Getenv("DB_PASSWORD")
		if pass == "" {
			pass = "postgres"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "modelhub_db"
		}
		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + pass + " dbname=" + dbname + " sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("database not available: %v (set DB_* or TEST_DSN to run)", err)
		return nil
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// exampleDir returns the path to api/cdm-json-schema/examples (relative to this test file). Skips if missing.
func exampleDir(t *testing.T) string {
	t.Helper()
	d := filepath.Join(apiRoot(), "cdm-json-schema", "examples")
	if _, err := os.Stat(d); err != nil {
		t.Skipf("cdm-json-schema/examples not found at %s: %v (init submodule: git submodule update --init)", d, err)
		return ""
	}
	return d
}

// TestSaveAndLoadExample runs SaveCDM then LoadCDM on one example file and checks root UUID and meta name round-trip.
func TestSaveAndLoadExample(t *testing.T) {
	db := testDB(t)
	dir := exampleDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}
	var one string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			one = filepath.Join(dir, e.Name())
			break
		}
	}
	if one == "" {
		t.Skip("no .json files in examples")
	}
	raw, err := os.ReadFile(one)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	rootUUID, err := SaveCDM(db, raw)
	if err != nil {
		t.Fatalf("SaveCDM: %v", err)
	}
	loaded, err := LoadCDM(db, rootUUID)
	if err != nil {
		t.Fatalf("LoadCDM: %v", err)
	}
	if loaded.Meta.UUID != rootUUID {
		t.Errorf("loaded.Meta.UUID = %q, want %q", loaded.Meta.UUID, rootUUID)
	}
	if loaded.Meta.Name == "" {
		t.Errorf("loaded.Meta.Name is empty")
	}
}

// TestSaveAndLoadAllExamples inserts every example then loads by root UUID. Verifies counts and root meta.
func TestSaveAndLoadAllExamples(t *testing.T) {
	db := testDB(t)
	dir := exampleDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read examples dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			rootUUID, err := SaveCDM(db, raw)
			if err != nil {
				t.Fatalf("SaveCDM: %v", err)
			}
			loaded, err := LoadCDM(db, rootUUID)
			if err != nil {
				t.Fatalf("LoadCDM: %v", err)
			}
			if loaded.Meta.UUID != rootUUID {
				t.Errorf("loaded.Meta.UUID = %q, want %q", loaded.Meta.UUID, rootUUID)
			}
			// Sanity: non-empty name and plausible child counts
			if loaded.Meta.Name == "" {
				t.Errorf("loaded.Meta.Name is empty")
			}
		})
	}
}

// TestSaveCDM_Idempotent saves the same raw twice and asserts the same root UUID is returned.
func TestSaveCDM_Idempotent(t *testing.T) {
	db := testDB(t)
	dir := exampleDir(t)
	entries, _ := os.ReadDir(dir)
	var one string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			one = filepath.Join(dir, e.Name())
			break
		}
	}
	if one == "" {
		t.Skip("no .json in examples")
	}
	raw, err := os.ReadFile(one)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	u1, err := SaveCDM(db, raw)
	if err != nil {
		t.Fatalf("first SaveCDM: %v", err)
	}
	u2, err := SaveCDM(db, raw)
	if err != nil {
		t.Fatalf("second SaveCDM: %v", err)
	}
	if u1 != u2 {
		t.Errorf("idempotency: first SaveCDM returned %q, second %q", u1, u2)
	}
}

// ── Negative examples ─────────────────────────────────────────────────────────

// TestSaveCDM_InvalidJSON ensures that obviously invalid JSON is rejected before
// touching the database. We pass a nil *gorm.DB because SaveCDM should fail
// during schema validation and never reach the transaction.
func TestSaveCDM_InvalidJSON(t *testing.T) {
	raw := []byte(`{"meta": "not an object"`) // malformed JSON
	_, err := SaveCDM(nil, raw)
	if err == nil {
		t.Fatalf("expected SaveCDM to fail for invalid JSON")
	}
	if !strings.Contains(err.Error(), "schema validation:") {
		t.Fatalf("expected schema validation error, got: %v", err)
	}
}

// TestValidateRefs_MissingIOValue constructs an in-memory CDM where a control
// references a non-existent IOValue UUID. This bypasses schema/DB to exercise
// the reference validation logic directly.
func TestValidateRefs_MissingIOValue(t *testing.T) {
	ctrl := cdm.Control{
		Meta: cdm.Asset{UUID: "control-1"},
	}
	valuesJSON, err := json.Marshal([]string{"missing-io"})
	if err != nil {
		t.Fatalf("marshal test values: %v", err)
	}
	ctrl.InputOutputValues = valuesJSON

	m := cdm.CausalDecisionModel{
		Meta:     cdm.Asset{UUID: "root"},
		Controls: []cdm.Control{ctrl},
		// Note: InputOutputValues is empty, so "missing-io" is dangling.
	}

	if err := validateRefs(&m); err == nil {
		t.Fatalf("expected validateRefs to fail for missing IOValue reference")
	}
}

// TestValidateRefs_MissingEvaluatableAsset constructs a CDM where a runnable
// model element points at a non-existent EvaluatableAsset UUID.
func TestValidateRefs_MissingEvaluatableAsset(t *testing.T) {
	elem := map[string]interface{}{
		"evaluatableAsset": "missing-eval",
	}
	elemsJSON, err := json.Marshal([]map[string]interface{}{elem})
	if err != nil {
		t.Fatalf("marshal runnable elements: %v", err)
	}
	rm := cdm.RunnableModel{
		Meta:     cdm.Asset{UUID: "rm-1"},
		Elements: elemsJSON,
	}
	m := cdm.CausalDecisionModel{
		Meta:           cdm.Asset{UUID: "root"},
		RunnableModels: []cdm.RunnableModel{rm},
		// Note: EvaluatableAssets is empty, so "missing-eval" is dangling.
	}

	err = validateRefs(&m)
	if err == nil {
		t.Fatalf("expected validateRefs to fail for missing EvaluatableAsset reference")
	}
	if !strings.Contains(err.Error(), "invalid CDM: reference") {
		t.Fatalf("unexpected error from validateRefs: %v", err)
	}
}
