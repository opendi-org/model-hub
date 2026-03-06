package cdm

import "encoding/json"

// Diagram (DI-Diagram) is a diagram component of a CDM.
// Elements, dependencies, and addons are JSONB blobs; validated by schema at boundary.
type Diagram struct {
	Meta         Asset           `json:"meta"`
	Elements     json.RawMessage `json:"elements,omitempty"`
	Dependencies json.RawMessage `json:"dependencies,omitempty"`
	Addons       json.RawMessage `json:"addons,omitempty"`
}
