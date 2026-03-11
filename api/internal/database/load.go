package database

// load.go — reconstructs a CausalDecisionModel from PostgreSQL.
//
// LoadCDM uses batched IN queries (one per child type) instead of
// per-row lookups, avoiding the N+1 problem present in the original
// converter.go implementation.

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// LoadCDM fetches the CDM identified by rootUUID and reconstructs the full
// domain model. Returns gorm.ErrRecordNotFound if the UUID does not exist.
func LoadCDM(db *gorm.DB, rootUUID string) (*cdm.CausalDecisionModel, error) {
	// ── Root row ──────────────────────────────────────────────────────────────
	var root CDMModel
	if err := db.Where("meta_uuid = ?", rootUUID).First(&root).Error; err != nil {
		return nil, fmt.Errorf("loading cdm_model %s: %w", rootUUID, err)
	}

	out := &cdm.CausalDecisionModel{
		Schema: root.Schema,
		Meta:   fieldsToMeta(root.Meta),
		Addons: json.RawMessage(root.Addons),
	}

	// ── IOValues ──────────────────────────────────────────────────────────────
	ioUUIDs, err := childUUIDs[CDMModelIOValue](db, rootUUID)
	if err != nil {
		return nil, fmt.Errorf("loading io_value links for %s: %w", rootUUID, err)
	}
	if len(ioUUIDs) > 0 {
		var rows []CDMIOValue
		if err := db.Where("meta_uuid IN ?", ioUUIDs).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("loading io_values for %s: %w", rootUUID, err)
		}
		out.InputOutputValues = make([]cdm.IOValue, len(rows))
		for i, r := range rows {
			out.InputOutputValues[i] = cdm.IOValue{
				Meta: fieldsToMeta(r.Meta),
				Data: json.RawMessage(r.Data),
			}
		}
	}

	// ── EvaluatableAssets ─────────────────────────────────────────────────────
	evalUUIDs, err := childUUIDs[CDMModelEvaluatableAsset](db, rootUUID)
	if err != nil {
		return nil, fmt.Errorf("loading evaluatable_asset links for %s: %w", rootUUID, err)
	}
	if len(evalUUIDs) > 0 {
		var rows []CDMEvaluatableAsset
		if err := db.Where("meta_uuid IN ?", evalUUIDs).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("loading evaluatable_assets for %s: %w", rootUUID, err)
		}
		out.EvaluatableAssets = make([]cdm.EvaluatableAsset, len(rows))
		for i, r := range rows {
			out.EvaluatableAssets[i] = cdm.EvaluatableAsset{
				Meta:     fieldsToMeta(r.Meta),
				EvalType: r.EvalType,
				Content:  json.RawMessage(r.Content),
			}
		}
	}

	// ── Diagrams ──────────────────────────────────────────────────────────────
	diagUUIDs, err := childUUIDs[CDMModelDiagram](db, rootUUID)
	if err != nil {
		return nil, fmt.Errorf("loading diagram links for %s: %w", rootUUID, err)
	}
	if len(diagUUIDs) > 0 {
		var rows []CDMDiagram
		if err := db.Where("meta_uuid IN ?", diagUUIDs).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("loading diagrams for %s: %w", rootUUID, err)
		}
		out.Diagrams = make([]cdm.Diagram, len(rows))
		for i, r := range rows {
			out.Diagrams[i] = cdm.Diagram{
				Meta:         fieldsToMeta(r.Meta),
				Elements:     json.RawMessage(r.Elements),
				Dependencies: json.RawMessage(r.Dependencies),
				Addons:       json.RawMessage(r.Addons),
			}
		}
	}

	// ── RunnableModels ────────────────────────────────────────────────────────
	runnableUUIDs, err := childUUIDs[CDMModelRunnableModel](db, rootUUID)
	if err != nil {
		return nil, fmt.Errorf("loading runnable_model links for %s: %w", rootUUID, err)
	}
	if len(runnableUUIDs) > 0 {
		var rows []CDMRunnableModel
		if err := db.Where("meta_uuid IN ?", runnableUUIDs).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("loading runnable_models for %s: %w", rootUUID, err)
		}
		out.RunnableModels = make([]cdm.RunnableModel, len(rows))
		for i, r := range rows {
			out.RunnableModels[i] = cdm.RunnableModel{
				Meta:     fieldsToMeta(r.Meta),
				Elements: json.RawMessage(r.Elements),
				Addons:   json.RawMessage(r.Addons),
			}
		}
	}

	// ── Controls ──────────────────────────────────────────────────────────────
	controlUUIDs, err := childUUIDs[CDMModelControl](db, rootUUID)
	if err != nil {
		return nil, fmt.Errorf("loading control links for %s: %w", rootUUID, err)
	}
	if len(controlUUIDs) > 0 {
		var rows []CDMControl
		if err := db.Where("meta_uuid IN ?", controlUUIDs).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("loading controls for %s: %w", rootUUID, err)
		}
		out.Controls = make([]cdm.Control, len(rows))
		for i, r := range rows {
			out.Controls[i] = cdm.Control{
				Meta:              fieldsToMeta(r.Meta),
				InputOutputValues: json.RawMessage(r.InputOutputValues),
				Displays:          json.RawMessage(r.Displays),
			}
		}
	}

	return out, nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

// linkRow is implemented by every CDMModel* join type so we can extract the
// child UUID without reflection or JSON round-trips.
type linkRow interface {
	childUUID() string
}

func (r CDMModelIOValue) childUUID() string          { return r.IOValueUUID }
func (r CDMModelEvaluatableAsset) childUUID() string { return r.EvaluatableUUID }
func (r CDMModelDiagram) childUUID() string          { return r.DiagramUUID }
func (r CDMModelRunnableModel) childUUID() string    { return r.RunnableUUID }
func (r CDMModelControl) childUUID() string          { return r.ControlUUID }

// childUUIDs fetches every child UUID from join table T for the given model.
// Returns nil (not an error) when no links exist.
func childUUIDs[T linkRow](db *gorm.DB, rootUUID string) ([]string, error) {
	var links []T
	if err := db.Where("model_uuid = ?", rootUUID).Find(&links).Error; err != nil {
		return nil, err
	}
	uuids := make([]string, 0, len(links))
	for _, l := range links {
		if u := l.childUUID(); u != "" {
			uuids = append(uuids, u)
		}
	}
	return uuids, nil
}
