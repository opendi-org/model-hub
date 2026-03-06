package cdm

import "encoding/json"

// Control (DI-Control) connects I/O values to diagram displays.
// InputOutputValues and displays are UUID arrays stored as JSONB.
type Control struct {
	Meta              Asset           `json:"meta"`
	InputOutputValues json.RawMessage `json:"inputOutputValues"`
	Displays          json.RawMessage `json:"displays"`
}
