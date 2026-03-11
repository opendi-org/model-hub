package database

// validate.go — two-phase validation for incoming CDM documents.
//
// Phase 1 (validateSchema):   structural/type checks via the OpenDI JSON Schema.
//                              Operates on raw []byte so nothing is silently
//                              dropped by an intermediate Go struct.
//
// Phase 2 (validateRefs):     logical cross-reference checks.
//                              Runs AFTER schema validation, BEFORE any UUID
//                              rewriting, so all UUIDs are still client-provided.
//
// Neither function mutates the CDM.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// ── Schema path ───────────────────────────────────────────────────────────────

// cdmSchemaPath is the path to the root CDM JSON Schema file.
// Override in tests by setting this before the first call to validateSchema.
var cdmSchemaPath = "./cdm-json-schema/schema-source/Causal-Decision-Model.json"

// schemaURLPrefix is the $id prefix used in OpenDI schema files.
// All $ref URIs starting with this prefix are resolved to local files.
const schemaURLPrefix = "https://opendi.org/json-schema/v0-draft-0325/"

var (
	compiledSchemaOnce sync.Once
	compiledSchema     *jsonschema.Schema
	compiledSchemaErr  error
)

func getCompiledSchema() (*jsonschema.Schema, error) {
	compiledSchemaOnce.Do(func() {
		baseDir := filepath.Dir(cdmSchemaPath)
		loader := jsonschema.SchemeURLLoader{
			"file":  jsonschema.FileLoader{},
			"https": localSchemaLoader{baseDir: baseDir},
		}
		compiler := jsonschema.NewCompiler()
		compiler.UseLoader(loader)
		compiledSchema, compiledSchemaErr = compiler.Compile(cdmSchemaPath)
	})
	return compiledSchema, compiledSchemaErr
}

// localSchemaLoader resolves opendi.org $id/$ref URLs to local schema files
// so that schema compilation never makes network requests.
type localSchemaLoader struct {
	baseDir string
}

func (l localSchemaLoader) Load(url string) (any, error) {
	if !strings.HasPrefix(url, schemaURLPrefix) {
		// Not an OpenDI URL — we don't handle it; let the caller error.
		return nil, fmt.Errorf("unsupported schema URL (not an opendi.org schema): %s", url)
	}
	name := strings.TrimPrefix(url, schemaURLPrefix)
	localPath := filepath.Join(l.baseDir, name)
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("opening local schema %q (resolved from %s): %w", localPath, url, err)
	}
	defer f.Close()
	return jsonschema.UnmarshalJSON(f)
}

// ── Phase 1: JSON Schema validation ──────────────────────────────────────────

// validateSchema validates raw CDM JSON bytes against the compiled OpenDI schema.
// Using raw bytes (not a re-marshalled struct) ensures that fields unknown to
// the Go model are not silently dropped before validation.
func validateSchema(raw []byte) error {
	schema, err := getCompiledSchema()
	if err != nil {
		return fmt.Errorf("loading CDM schema: %w", err)
	}
	v, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("parsing CDM JSON: %w", err)
	}
	if err := schema.Validate(v); err != nil {
		return fmt.Errorf("CDM schema validation failed: %w", err)
	}
	return nil
}

// ── Phase 2: Internal reference validation ────────────────────────────────────

// validateRefs checks that every UUID cross-reference inside the CDM points to
// an existing sibling. It runs on the parsed struct after schema validation but
// before any UUID rewriting, so all UUIDs are still the client-provided values.
//
// Checks performed:
//   - runnableModels[*].elements[*].evaluatableAsset  → must exist in evaluatableAssets
//   - runnableModels[*].elements[*].inputs[]          → must exist in inputOutputValues
//   - runnableModels[*].elements[*].outputs[]         → must exist in inputOutputValues
//   - controls[*].inputOutputValues[]                 → must exist in inputOutputValues
//   - controls[*].displays[]                          → must exist in diagrams[*].elements[*].displays[*].meta.uuid
func validateRefs(m *cdm.CausalDecisionModel) error {
	ioSet, evalSet, err := buildLeafSets(m)
	if err != nil {
		return err
	}
	displaySet, err := buildDisplaySet(m)
	if err != nil {
		return err
	}

	if err := checkRunnableRefs(m, ioSet, evalSet); err != nil {
		return err
	}
	if err := checkControlRefs(m, ioSet, displaySet); err != nil {
		return err
	}
	return nil
}

// buildLeafSets returns sets of known IOValue and EvaluatableAsset UUIDs.
func buildLeafSets(m *cdm.CausalDecisionModel) (ioSet, evalSet map[string]struct{}, err error) {
	ioSet = make(map[string]struct{}, len(m.InputOutputValues))
	for _, v := range m.InputOutputValues {
		if v.Meta.UUID != "" {
			ioSet[v.Meta.UUID] = struct{}{}
		}
	}
	evalSet = make(map[string]struct{}, len(m.EvaluatableAssets))
	for _, e := range m.EvaluatableAssets {
		if e.Meta.UUID != "" {
			evalSet[e.Meta.UUID] = struct{}{}
		}
	}
	return ioSet, evalSet, nil
}

// buildDisplaySet collects every display meta.uuid from all diagram elements.
// Returns an error if any display is missing its meta.uuid (schema guarantees
// this won't happen after validateSchema passes, but we check defensively).
func buildDisplaySet(m *cdm.CausalDecisionModel) (map[string]struct{}, error) {
	displaySet := make(map[string]struct{})
	for di, d := range m.Diagrams {
		if len(d.Elements) == 0 {
			continue
		}
		var elems []map[string]interface{}
		if err := json.Unmarshal(d.Elements, &elems); err != nil {
			return nil, fmt.Errorf("diagrams[%d].elements: invalid JSON: %w", di, err)
		}
		for ei, elem := range elems {
			rawDisplays, _ := elem["displays"].([]interface{})
			for diIdx, disp := range rawDisplays {
				dispMap, ok := disp.(map[string]interface{})
				if !ok {
					continue
				}
				meta, _ := dispMap["meta"].(map[string]interface{})
				u, _ := meta["uuid"].(string)
				if u == "" {
					return nil, fmt.Errorf(
						"diagrams[%d].elements[%d].displays[%d]: missing meta.uuid", di, ei, diIdx,
					)
				}
				displaySet[u] = struct{}{}
			}
		}
	}
	return displaySet, nil
}

// checkRunnableRefs validates evaluatableAsset, inputs, and outputs references
// inside every runnable model's elements array.
func checkRunnableRefs(m *cdm.CausalDecisionModel, ioSet, evalSet map[string]struct{}) error {
	for ri, rm := range m.RunnableModels {
		if len(rm.Elements) == 0 {
			continue
		}
		var elems []map[string]interface{}
		if err := json.Unmarshal(rm.Elements, &elems); err != nil {
			return fmt.Errorf("runnableModels[%d].elements: invalid JSON: %w", ri, err)
		}
		for ei, elem := range elems {
			if v, _ := elem["evaluatableAsset"].(string); v != "" {
				if _, ok := evalSet[v]; !ok {
					return refErr("runnableModels[%d].elements[%d].evaluatableAsset", ri, ei, v)
				}
			}
			if err := checkUUIDArrayRefs(elem, "inputs", ioSet,
				fmt.Sprintf("runnableModels[%d].elements[%d]", ri, ei)); err != nil {
				return err
			}
			if err := checkUUIDArrayRefs(elem, "outputs", ioSet,
				fmt.Sprintf("runnableModels[%d].elements[%d]", ri, ei)); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkControlRefs validates inputOutputValues and displays references on every control.
func checkControlRefs(m *cdm.CausalDecisionModel, ioSet, displaySet map[string]struct{}) error {
	for ci, ctrl := range m.Controls {
		if len(ctrl.InputOutputValues) > 0 {
			var arr []interface{}
			if err := json.Unmarshal(ctrl.InputOutputValues, &arr); err != nil {
				return fmt.Errorf("controls[%d].inputOutputValues: invalid JSON: %w", ci, err)
			}
			for idx, v := range arr {
				if s, _ := v.(string); s != "" {
					if _, ok := ioSet[s]; !ok {
						return refErr("controls[%d].inputOutputValues[%d]", ci, idx, s)
					}
				}
			}
		}

		// Always check display refs, even if displaySet is empty —
		// any reference with an empty display set is by definition dangling.
		if len(ctrl.Displays) > 0 {
			var arr []interface{}
			if err := json.Unmarshal(ctrl.Displays, &arr); err != nil {
				return fmt.Errorf("controls[%d].displays: invalid JSON: %w", ci, err)
			}
			for idx, v := range arr {
				if s, _ := v.(string); s != "" {
					if _, ok := displaySet[s]; !ok {
						return refErr("controls[%d].displays[%d]", ci, idx, s)
					}
				}
			}
		}
	}
	return nil
}

// checkUUIDArrayRefs validates that every UUID in elem[field] exists in the given set.
func checkUUIDArrayRefs(elem map[string]interface{}, field string, set map[string]struct{}, prefix string) error {
	arr, _ := elem[field].([]interface{})
	for idx, x := range arr {
		if s, _ := x.(string); s != "" {
			if _, ok := set[s]; !ok {
				return fmt.Errorf("invalid CDM: reference %q not found for %s.%s[%d]", s, prefix, field, idx)
			}
		}
	}
	return nil
}

// refErr formats a dangling-reference error with a printf-style path.
func refErr(pathFmt string, a, b int, uuid string) error {
	path := fmt.Sprintf(pathFmt, a, b)
	return fmt.Errorf("invalid CDM: reference %q not found for %s", uuid, path)
}
