package cdm

import "encoding/json"

// Asset (DI-Asset) is the base metadata shared by all DI assets.
// Every "meta" in the schema is this type.
// Documentation is stored as JSONB; shape validated by schema at boundary.
type Asset struct {
	UUID          string          `json:"uuid"`
	Name          string          `json:"name,omitempty"`
	Summary       string          `json:"summary,omitempty"`
	Documentation json.RawMessage `json:"documentation,omitempty"`
	Version       string          `json:"version,omitempty"`
	Draft         *bool           `json:"draft,omitempty"`
	Creator       UUID            `json:"creator,omitempty"`
	CreatedDate   string          `json:"createdDate,omitempty"`
	Updator       UUID            `json:"updator,omitempty"`
	UpdatedDate   string          `json:"updatedDate,omitempty"`
}
