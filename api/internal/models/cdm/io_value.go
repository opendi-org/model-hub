package cdm

import "encoding/json"

// IOValue (DI-IO-Value) represents an I/O value used by runnable models and controls.
// Data can be any JSON type per schema.
type IOValue struct {
	Meta Asset `json:"meta"`
	Data json.RawMessage `json:"data"`
}
