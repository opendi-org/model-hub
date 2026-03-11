package cdm

import "encoding/json"

// RunnableModel (DI-Runnable-Model) is a runnable component of a CDM.
// Elements and addons are JSONB blobs; validated by schema at boundary.
type RunnableModel struct {
	Meta     Asset           `json:"meta"`
	Elements json.RawMessage `json:"elements"`
	Addons   json.RawMessage `json:"addons,omitempty"`
}
