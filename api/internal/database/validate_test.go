package database

import (
	"strings"
	"testing"
)

func TestValidateCDMWithIssues_SchemaProvidesPathAndLine(t *testing.T) {
	raw := []byte(`{
  "meta": {
    "uuid": "not-a-valid-uuid",
    "name": 42,
    "version": false
  },
  "runnableModels": [],
  "diagrams": [],
  "controls": [],
  "inputOutputValues": [],
  "evaluatableAssets": []
}`)

	issues, err := ValidateCDMWithIssues(raw)
	if err != nil {
		t.Fatalf("ValidateCDMWithIssues returned error: %v", err)
	}
	if len(issues) == 0 {
		t.Fatalf("expected schema issues, got none")
	}

	found := false
	for _, is := range issues {
		if is.InstancePath == "/meta/uuid" {
			found = true
			if is.Line <= 0 || is.Column <= 0 {
				t.Fatalf("expected line/column for /meta/uuid issue, got line=%d col=%d", is.Line, is.Column)
			}
			if is.Phase != "schema" {
				t.Fatalf("expected phase=schema, got %q", is.Phase)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected an issue for /meta/uuid; got: %+v", issues)
	}
}

func TestPathExprToPointer(t *testing.T) {
	got := pathExprToPointer("controls[2].displays[1]")
	want := "/controls/2/displays/1"
	if got != want {
		t.Fatalf("pathExprToPointer mismatch: got %q want %q", got, want)
	}
}

func TestReferenceValidationIssue_ParsesPath(t *testing.T) {
	raw := []byte(`{
  "controls": [
    { "displays": ["x"] }
  ]
}`)
	err := refErr("controls[%d].displays[%d]", 0, 0, "abc")
	issue := referenceValidationIssue(raw, err)
	if issue.Phase != "reference" {
		t.Fatalf("expected reference phase, got %q", issue.Phase)
	}
	if issue.InstancePath != "/controls/0/displays/0" {
		t.Fatalf("unexpected instance path: %q", issue.InstancePath)
	}
	if !strings.Contains(issue.Message, "reference") {
		t.Fatalf("unexpected message: %q", issue.Message)
	}
}
