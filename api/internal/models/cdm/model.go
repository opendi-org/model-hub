package cdm

import "encoding/json"

// CausalDecisionModel is the root CDM document (full schema).
type CausalDecisionModel struct {
	Schema            string             `json:"$schema,omitempty"`
	Meta              Asset              `json:"meta"`
	RunnableModels    []RunnableModel    `json:"runnableModels,omitempty"`
	Diagrams          []Diagram          `json:"diagrams,omitempty"`
	EvaluatableAssets []EvaluatableAsset `json:"evaluatableAssets,omitempty"`
	InputOutputValues []IOValue          `json:"inputOutputValues,omitempty"`
	Controls          []Control          `json:"controls,omitempty"`
	Addons            json.RawMessage    `json:"addons,omitempty"`
}
