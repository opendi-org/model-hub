package database

import (
	"encoding/json"
	"testing"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// canonicalizeJSON 

// TestCanonicalizeJSON_KeyOrder verifies that two JSON objects with the same
// keys in different order produce identical bytes after canonicalization.
func TestCanonicalizeJSON_KeyOrder(t *testing.T) {
	a := json.RawMessage(`{"b":1,"a":2}`)
	b := json.RawMessage(`{"a":2,"b":1}`)

	ca, err := canonicalizeJSON(a)
	if err != nil {
		t.Fatalf("canonicalizeJSON a: %v", err)
	}
	cb, err := canonicalizeJSON(b)
	if err != nil {
		t.Fatalf("canonicalizeJSON b: %v", err)
	}
	if string(ca) != string(cb) {
		t.Errorf("expected same bytes after canonicalize: %q vs %q", ca, cb)
	}
}

// TestCanonicalizeJSON_Empty verifies that empty input is returned unchanged.
func TestCanonicalizeJSON_Empty(t *testing.T) {
	out, err := canonicalizeJSON(nil)
	if err != nil {
		t.Fatalf("canonicalizeJSON nil: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output, got %q", out)
	}
}

//  sha256UUID 

// TestSha256UUID_Stable verifies that the same input always produces the same UUID.
func TestSha256UUID_Stable(t *testing.T) {
	input := []byte("hello world")
	u1 := sha256UUID(input)
	u2 := sha256UUID(input)
	if u1 != u2 {
		t.Errorf("sha256UUID not stable: %q vs %q", u1, u2)
	}
}

// TestSha256UUID_DifferentInput verifies that different inputs produce different UUIDs.
func TestSha256UUID_DifferentInput(t *testing.T) {
	u1 := sha256UUID([]byte("content A"))
	u2 := sha256UUID([]byte("content B"))
	if u1 == u2 {
		t.Errorf("expected different UUIDs for different inputs, got %q", u1)
	}
}

// TestSha256UUID_Format verifies the output looks like a UUID (8-4-4-4-12).
func TestSha256UUID_Format(t *testing.T) {
	u := sha256UUID([]byte("test"))
	if len(u) != 36 {
		t.Errorf("expected UUID length 36, got %d: %q", len(u), u)
	}
	if u[8] != '-' || u[13] != '-' || u[18] != '-' || u[23] != '-' {
		t.Errorf("expected UUID format 8-4-4-4-12, got %q", u)
	}
}

// Same content same UUID
// TestAssignDigests_SameContentSameUUID verifies that two IOValues with identical
// content but different client UUIDs receive the same content-addressed UUID.
func TestAssignDigests_SameContentSameUUID(t *testing.T) {
	data := json.RawMessage(`{"value":42}`)

	m := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "client-uuid-1", Name: "temp"}, Data: data},
			{Meta: cdm.Asset{UUID: "client-uuid-2", Name: "temp"}, Data: data},
		},
	}

	if _, err := assignDigests(&m); err != nil {
		t.Fatalf("assignDigests: %v", err)
	}

	uuid1 := m.InputOutputValues[0].Meta.UUID
	uuid2 := m.InputOutputValues[1].Meta.UUID
	if uuid1 != uuid2 {
		t.Errorf("identical IOValues got different UUIDs: %q vs %q", uuid1, uuid2)
	}
}

// TestAssignDigests_DifferentContentDifferentUUID verifies that IOValues with
// different data produce different UUIDs.
func TestAssignDigests_DifferentContentDifferentUUID(t *testing.T) {
	m := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c1", Name: "temp"}, Data: json.RawMessage(`{"value":1}`)},
			{Meta: cdm.Asset{UUID: "c2", Name: "temp"}, Data: json.RawMessage(`{"value":2}`)},
		},
	}

	if _, err := assignDigests(&m); err != nil {
		t.Fatalf("assignDigests: %v", err)
	}

	uuid1 := m.InputOutputValues[0].Meta.UUID
	uuid2 := m.InputOutputValues[1].Meta.UUID
	if uuid1 == uuid2 {
		t.Errorf("different IOValues got same UUID: %q", uuid1)
	}
}

//  Hub fields don't affect the hash 

// TestAssignDigests_HubFieldsIgnored verifies that creator/createdDate/updator/
// updatedDate do not affect the content-addressed UUID.
func TestAssignDigests_HubFieldsIgnored(t *testing.T) {
	data := json.RawMessage(`{"value":99}`)

	withHub := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{
				Meta: cdm.Asset{
					UUID:        "c1",
					Name:        "temp",
					Creator:     "john",
					CreatedDate: "2024-01-01",
					Updator:     "jane",
					UpdatedDate: "2024-06-01",
				},
				Data: data,
			},
		},
	}
	withoutHub := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c2", Name: "temp"}, Data: data},
		},
	}

	if _, err := assignDigests(&withHub); err != nil {
		t.Fatalf("assignDigests withHub: %v", err)
	}
	if _, err := assignDigests(&withoutHub); err != nil {
		t.Fatalf("assignDigests withoutHub: %v", err)
	}

	if withHub.InputOutputValues[0].Meta.UUID != withoutHub.InputOutputValues[0].Meta.UUID {
		t.Errorf("hub fields changed the UUID: %q vs %q",
			withHub.InputOutputValues[0].Meta.UUID,
			withoutHub.InputOutputValues[0].Meta.UUID)
	}
}

//Root UUID is stable 

// TestAssignDigests_RootUUIDStable verifies that calling assignDigests twice on
// identical input produces the same root UUID.
func TestAssignDigests_RootUUIDStable(t *testing.T) {
	newModel := func() cdm.CausalDecisionModel {
		return cdm.CausalDecisionModel{
			Meta: cdm.Asset{UUID: "root", Name: "my model"},
			InputOutputValues: []cdm.IOValue{
				{Meta: cdm.Asset{UUID: "c1", Name: "temp"}, Data: json.RawMessage(`{"value":1}`)},
			},
		}
	}

	m1 := newModel()
	m2 := newModel()

	r1, err := assignDigests(&m1)
	if err != nil {
		t.Fatalf("assignDigests first: %v", err)
	}
	r2, err := assignDigests(&m2)
	if err != nil {
		t.Fatalf("assignDigests second: %v", err)
	}

	if r1 != r2 {
		t.Errorf("root UUID not stable: %q vs %q", r1, r2)
	}
}

// TestAssignDigests_RootUUIDChangesWithContent verifies that changing a child
// asset's content changes the root UUID.
func TestAssignDigests_RootUUIDChangesWithContent(t *testing.T) {
	m1 := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "my model"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c1", Name: "temp"}, Data: json.RawMessage(`{"value":1}`)},
		},
	}
	m2 := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "my model"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c1", Name: "temp"}, Data: json.RawMessage(`{"value":2}`)},
		},
	}

	r1, err := assignDigests(&m1)
	if err != nil {
		t.Fatalf("assignDigests m1: %v", err)
	}
	r2, err := assignDigests(&m2)
	if err != nil {
		t.Fatalf("assignDigests m2: %v", err)
	}

	if r1 == r2 {
		t.Errorf("expected different root UUIDs for different content, got %q", r1)
	}
}

// Reference rewriting 

// TestAssignDigests_RunnableRefsRewritten verifies that after assignDigests, a
// RunnableModel's elements blob has its IOValue/EvaluatableAsset references
// updated to the new content-addressed UUIDs.
func TestAssignDigests_RunnableRefsRewritten(t *testing.T) {
	// Build a RunnableModel that references an IOValue by client UUID "old-io-1"
	// and an EvaluatableAsset by client UUID "old-eval-1".
	elements := json.RawMessage(`[{"inputs":["old-io-1"],"outputs":[],"evaluatableAsset":"old-eval-1"}]`)

	m := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "old-io-1", Name: "temp"}, Data: json.RawMessage(`{"v":1}`)},
		},
		EvaluatableAssets: []cdm.EvaluatableAsset{
			{Meta: cdm.Asset{UUID: "old-eval-1", Name: "script"}, EvalType: "python", Content: json.RawMessage(`{"code":"print(1)"}`)},
		},
		RunnableModels: []cdm.RunnableModel{
			{Meta: cdm.Asset{UUID: "rm-1", Name: "rm"}, Elements: elements},
		},
	}

	if _, err := assignDigests(&m); err != nil {
		t.Fatalf("assignDigests: %v", err)
	}

	newIOUUID := m.InputOutputValues[0].Meta.UUID
	newEvalUUID := m.EvaluatableAssets[0].Meta.UUID

	// Neither should still be the old client UUID
	if newIOUUID == "old-io-1" {
		t.Errorf("IOValue UUID was not rewritten")
	}
	if newEvalUUID == "old-eval-1" {
		t.Errorf("EvaluatableAsset UUID was not rewritten")
	}

	// The RunnableModel's elements blob should now reference the new UUIDs
	var elems []map[string]interface{}
	if err := json.Unmarshal(m.RunnableModels[0].Elements, &elems); err != nil {
		t.Fatalf("unmarshal elements: %v", err)
	}

	inputs, _ := elems[0]["inputs"].([]interface{})
	if len(inputs) == 0 || inputs[0].(string) != newIOUUID {
		t.Errorf("inputs[0] not rewritten: got %v, want %q", inputs, newIOUUID)
	}
	evalRef, _ := elems[0]["evaluatableAsset"].(string)
	if evalRef != newEvalUUID {
		t.Errorf("evaluatableAsset not rewritten: got %q, want %q", evalRef, newEvalUUID)
	}
}

// TestAssignDigests_ControlRefsRewritten verifies that a Control's
// inputOutputValues array is updated to the new content-addressed IOValue UUIDs.
func TestAssignDigests_ControlRefsRewritten(t *testing.T) {
	iovJSON, _ := json.Marshal([]string{"old-io-1"})

	m := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "old-io-1", Name: "temp"}, Data: json.RawMessage(`{"v":1}`)},
		},
		Controls: []cdm.Control{
			{
				Meta:              cdm.Asset{UUID: "ctrl-1", Name: "ctrl"},
				InputOutputValues: iovJSON,
			},
		},
	}

	if _, err := assignDigests(&m); err != nil {
		t.Fatalf("assignDigests: %v", err)
	}

	newIOUUID := m.InputOutputValues[0].Meta.UUID
	if newIOUUID == "old-io-1" {
		t.Errorf("IOValue UUID was not rewritten")
	}

	var refs []string
	if err := json.Unmarshal(m.Controls[0].InputOutputValues, &refs); err != nil {
		t.Fatalf("unmarshal control IOValues: %v", err)
	}
	if len(refs) == 0 || refs[0] != newIOUUID {
		t.Errorf("control IOValue ref not rewritten: got %v, want %q", refs, newIOUUID)
	}
}

// TestAssignDigests_ChildOrderIrrelevant verifies that two CDMs with identical
// children in different insertion order produce the same root UUID.
// This exercises the sort.Slice in hashModel.
func TestAssignDigests_ChildOrderIrrelevant(t *testing.T) {
	ioA := cdm.IOValue{Meta: cdm.Asset{UUID: "c-a", Name: "a"}, Data: json.RawMessage(`{"v":1}`)}
	ioB := cdm.IOValue{Meta: cdm.Asset{UUID: "c-b", Name: "b"}, Data: json.RawMessage(`{"v":2}`)}
	d1 := cdm.Diagram{Meta: cdm.Asset{UUID: "d-1", Name: "d1"}}
	d2 := cdm.Diagram{Meta: cdm.Asset{UUID: "d-2", Name: "d2"}}

	m1 := cdm.CausalDecisionModel{
		Meta:              cdm.Asset{UUID: "root", Name: "model"},
		InputOutputValues: []cdm.IOValue{ioA, ioB},
		Diagrams:          []cdm.Diagram{d1, d2},
	}
	m2 := cdm.CausalDecisionModel{
		Meta:              cdm.Asset{UUID: "root", Name: "model"},
		InputOutputValues: []cdm.IOValue{ioB, ioA},
		Diagrams:          []cdm.Diagram{d2, d1},
	}

	r1, err := assignDigests(&m1)
	if err != nil {
		t.Fatalf("assignDigests m1: %v", err)
	}
	r2, err := assignDigests(&m2)
	if err != nil {
		t.Fatalf("assignDigests m2: %v", err)
	}

	if r1 != r2 {
		t.Errorf("child insertion order affected root UUID: %q vs %q", r1, r2)
	}
}

// TestAssignDigests_JSONKeyOrderIrrelevant verifies that an IOValue whose data
// has keys in different order produces the same UUID as one with keys sorted.
func TestAssignDigests_JSONKeyOrderIrrelevant(t *testing.T) {
	m1 := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c1", Name: "temp"}, Data: json.RawMessage(`{"b":2,"a":1}`)},
		},
	}
	m2 := cdm.CausalDecisionModel{
		Meta: cdm.Asset{UUID: "root", Name: "test"},
		InputOutputValues: []cdm.IOValue{
			{Meta: cdm.Asset{UUID: "c2", Name: "temp"}, Data: json.RawMessage(`{"a":1,"b":2}`)},
		},
	}

	if _, err := assignDigests(&m1); err != nil {
		t.Fatalf("assignDigests m1: %v", err)
	}
	if _, err := assignDigests(&m2); err != nil {
		t.Fatalf("assignDigests m2: %v", err)
	}

	if m1.InputOutputValues[0].Meta.UUID != m2.InputOutputValues[0].Meta.UUID {
		t.Errorf("JSON key order affected UUID: %q vs %q",
			m1.InputOutputValues[0].Meta.UUID,
			m2.InputOutputValues[0].Meta.UUID)
	}
}
