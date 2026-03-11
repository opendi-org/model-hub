package cdm

import "encoding/json"

// EvaluatableAsset is one evaluatable asset (script, apicall, etc.).
// Meta and evalType are explicit; content is raw JSON validated by schema.
type EvaluatableAsset struct {
	Meta     Asset           `json:"meta"`
	EvalType string          `json:"evalType"`
	Content  json.RawMessage `json:"content"`
}
