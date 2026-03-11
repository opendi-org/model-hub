package database

// store.go — persists a fully-validated, digest-assigned CDM to PostgreSQL.
//
// SaveCDM is the public entry point. It:
//   1. Validates raw JSON bytes against the OpenDI JSON Schema.
//   2. Parses into the domain model.
//   3. Validates internal UUID cross-references.
//   4. Canonicalizes JSONB blobs and assigns content-addressed UUIDs to all nodes.
//   5. Writes everything inside a single transaction (all-or-nothing).
//
// All content tables are append-only (INSERT ... ON CONFLICT DO NOTHING), so
// SaveCDM is fully idempotent: uploading the same content twice is safe.

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// SaveCDM validates, digests, and persists a CDM document.
//
// raw must be the original request body bytes — they are validated against the
// JSON Schema before any parsing, so nothing is lost through struct round-trip.
//
// Returns the content-addressed root UUID on success.
func SaveCDM(db *gorm.DB, raw []byte) (rootUUID string, err error) {
	// ── Phase 1: JSON Schema validation (on raw bytes) ────────────────────────
	if err := validateSchema(raw); err != nil {
		return "", fmt.Errorf("schema validation: %w", err)
	}

	// ── Phase 2: Parse ────────────────────────────────────────────────────────
	var m cdm.CausalDecisionModel
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", fmt.Errorf("parsing CDM: %w", err)
	}

	// ── Phase 3: Internal reference validation ────────────────────────────────
	// Run before UUID rewriting so all refs still carry client-provided values.
	if err := validateRefs(&m); err != nil {
		return "", fmt.Errorf("reference validation: %w", err)
	}

	// ── Phase 4: Canonicalize and assign content-addressed UUIDs ─────────────
	rootUUID, err = assignDigests(&m)
	if err != nil {
		return "", fmt.Errorf("digest assignment: %w", err)
	}

	// ── Phase 5: Persist inside a transaction ─────────────────────────────────
	now := time.Now().UTC()
	err = db.Transaction(func(tx *gorm.DB) error {
		return persistCDM(tx, &m, rootUUID, now)
	})
	if err != nil {
		return "", fmt.Errorf("persisting CDM %s: %w", rootUUID, err)
	}

	return rootUUID, nil
}

// persistCDM writes all rows for a single CDM document within an existing
// transaction. Every INSERT uses ON CONFLICT DO NOTHING so the operation is
// idempotent: re-uploading identical content is a no-op.
func persistCDM(tx *gorm.DB, m *cdm.CausalDecisionModel, rootUUID string, now time.Time) error {
	noop := clause.OnConflict{DoNothing: true}

	// ── Leaf content rows (order does not matter — no FK constraints) ─────────

	for _, iov := range m.InputOutputValues {
		row := CDMIOValue{
			CreatedAt: now,
			Meta:      metaToFields(iov.Meta),
			Data:      datatypes.JSON(iov.Data),
		}
		if err := tx.Clauses(noop).Create(&row).Error; err != nil {
			return fmt.Errorf("inserting io_value %s: %w", iov.Meta.UUID, err)
		}
	}

	for _, ea := range m.EvaluatableAssets {
		row := CDMEvaluatableAsset{
			CreatedAt: now,
			Meta:      metaToFields(ea.Meta),
			EvalType:  ea.EvalType,
			Content:   datatypes.JSON(ea.Content),
		}
		if err := tx.Clauses(noop).Create(&row).Error; err != nil {
			return fmt.Errorf("inserting evaluatable_asset %s: %w", ea.Meta.UUID, err)
		}
	}

	for _, dg := range m.Diagrams {
		row := CDMDiagram{
			CreatedAt:    now,
			Meta:         metaToFields(dg.Meta),
			Elements:     datatypes.JSON(dg.Elements),
			Dependencies: datatypes.JSON(dg.Dependencies),
			Addons:       datatypes.JSON(dg.Addons),
		}
		if err := tx.Clauses(noop).Create(&row).Error; err != nil {
			return fmt.Errorf("inserting diagram %s: %w", dg.Meta.UUID, err)
		}
	}

	for _, rm := range m.RunnableModels {
		row := CDMRunnableModel{
			CreatedAt: now,
			Meta:      metaToFields(rm.Meta),
			Elements:  datatypes.JSON(rm.Elements),
			Addons:    datatypes.JSON(rm.Addons),
		}
		if err := tx.Clauses(noop).Create(&row).Error; err != nil {
			return fmt.Errorf("inserting runnable_model %s: %w", rm.Meta.UUID, err)
		}
	}

	for _, ctrl := range m.Controls {
		row := CDMControl{
			CreatedAt:         now,
			Meta:              metaToFields(ctrl.Meta),
			InputOutputValues: datatypes.JSON(ctrl.InputOutputValues),
			Displays:          datatypes.JSON(ctrl.Displays),
		}
		if err := tx.Clauses(noop).Create(&row).Error; err != nil {
			return fmt.Errorf("inserting control %s: %w", ctrl.Meta.UUID, err)
		}
	}

	// ── Root CDM row ──────────────────────────────────────────────────────────

	root := CDMModel{
		CreatedAt: now,
		Schema:    m.Schema,
		Meta:      metaToFields(m.Meta),
		Addons:    datatypes.JSON(m.Addons),
	}
	if err := tx.Clauses(noop).Create(&root).Error; err != nil {
		return fmt.Errorf("inserting cdm_model %s: %w", rootUUID, err)
	}

	// ── Manifest join rows ────────────────────────────────────────────────────

	for _, iov := range m.InputOutputValues {
		link := CDMModelIOValue{CreatedAt: now, ModelUUID: rootUUID, IOValueUUID: iov.Meta.UUID}
		if err := tx.Clauses(noop).Create(&link).Error; err != nil {
			return fmt.Errorf("linking io_value %s: %w", iov.Meta.UUID, err)
		}
	}
	for _, ea := range m.EvaluatableAssets {
		link := CDMModelEvaluatableAsset{CreatedAt: now, ModelUUID: rootUUID, EvaluatableUUID: ea.Meta.UUID}
		if err := tx.Clauses(noop).Create(&link).Error; err != nil {
			return fmt.Errorf("linking evaluatable_asset %s: %w", ea.Meta.UUID, err)
		}
	}
	for _, dg := range m.Diagrams {
		link := CDMModelDiagram{CreatedAt: now, ModelUUID: rootUUID, DiagramUUID: dg.Meta.UUID}
		if err := tx.Clauses(noop).Create(&link).Error; err != nil {
			return fmt.Errorf("linking diagram %s: %w", dg.Meta.UUID, err)
		}
	}
	for _, rm := range m.RunnableModels {
		link := CDMModelRunnableModel{CreatedAt: now, ModelUUID: rootUUID, RunnableUUID: rm.Meta.UUID}
		if err := tx.Clauses(noop).Create(&link).Error; err != nil {
			return fmt.Errorf("linking runnable_model %s: %w", rm.Meta.UUID, err)
		}
	}
	for _, ctrl := range m.Controls {
		link := CDMModelControl{CreatedAt: now, ModelUUID: rootUUID, ControlUUID: ctrl.Meta.UUID}
		if err := tx.Clauses(noop).Create(&link).Error; err != nil {
			return fmt.Errorf("linking control %s: %w", ctrl.Meta.UUID, err)
		}
	}

	return nil
}
