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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// ValidationIssue is a machine-readable validation diagnostic.
// Used by HTTP handlers/clients to render actionable feedback.
type ValidationIssue struct {
	Phase        string `json:"phase"`                  // schema | reference
	InstancePath string `json:"instancePath"`           // JSON pointer-like path
	SchemaPath   string `json:"schemaPath,omitempty"`   // keyword path (schema phase)
	Message      string `json:"message"`                // human-readable issue text
	Line         int    `json:"line,omitempty"`         // 1-based line in source JSON
	Column       int    `json:"column,omitempty"`       // 1-based column in source JSON
}

// ── Schema path ───────────────────────────────────────────────────────────────

// cdmSchemaPath is the path to the root CDM JSON Schema file.
// Override in tests by setting this before the first call to validateSchema.
var cdmSchemaPath = defaultCDMSchemaPath()

func defaultCDMSchemaPath() string {
    _, thisFile, _, ok := runtime.Caller(0)
    if ok {
        p := filepath.Join(filepath.Dir(thisFile), "..", "..", "cdm-json-schema", "schema-source", "Causal-Decision-Model.json")
        if _, err := os.Stat(p); err == nil {
            return p
        }
    }
    return "./cdm-json-schema/schema-source/Causal-Decision-Model.json"
}

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

// ValidateCDM runs both validation phases (schema + refs) on raw CDM JSON bytes
// without persisting anything. Returns nil if the document is valid.
func ValidateCDM(raw []byte) error {
	if err := validateSchema(raw); err != nil {
		return fmt.Errorf("schema validation: %w", err)
	}
	var m cdm.CausalDecisionModel
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("parsing CDM: %w", err)
	}
	if err := validateRefs(&m); err != nil {
		return fmt.Errorf("reference validation: %w", err)
	}
	return nil
}

// ValidateCDMWithIssues runs schema+reference validation and returns structured
// issues for invalid documents. It returns (nil, nil) when valid.
// Returned error indicates validator infrastructure/processing failures.
func ValidateCDMWithIssues(raw []byte) ([]ValidationIssue, error) {
	schemaIssues, err := schemaValidationIssues(raw)
	if err != nil {
		return nil, fmt.Errorf("schema validation: %w", err)
	}
	if len(schemaIssues) > 0 {
		return schemaIssues, nil
	}

	var m cdm.CausalDecisionModel
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing CDM: %w", err)
	}
	if err := validateRefs(&m); err != nil {
		return []ValidationIssue{referenceValidationIssue(raw, err)}, nil
	}
	return nil, nil
}

func schemaValidationIssues(raw []byte) ([]ValidationIssue, error) {
	schema, err := getCompiledSchema()
	if err != nil {
		return nil, fmt.Errorf("loading CDM schema: %w", err)
	}
	v, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing CDM JSON: %w", err)
	}
	if err := schema.Validate(v); err != nil {
		var verr *jsonschema.ValidationError
		if !errors.As(err, &verr) {
			return nil, fmt.Errorf("CDM schema validation failed: %w", err)
		}
		pointerLC := buildJSONPointerLineMap(raw)
		out := verr.BasicOutput()
		issues := []ValidationIssue{}
		appendIssuesFromOutput(*out, pointerLC, &issues)
		if len(issues) == 0 {
			issues = append(issues, ValidationIssue{
				Phase:   "schema",
				Message: verr.Error(),
			})
		}
		return issues, nil
	}
	return nil, nil
}

func appendIssuesFromOutput(
	u jsonschema.OutputUnit,
	pointerLC map[string]lineCol,
	issues *[]ValidationIssue,
) {
	if u.Error != nil {
		issue := ValidationIssue{
			Phase:        "schema",
			InstancePath: u.InstanceLocation,
			SchemaPath:   u.KeywordLocation,
			Message:      u.Error.String(),
		}
		if lc, ok := pointerLC[issue.InstancePath]; ok {
			issue.Line = lc.line
			issue.Column = lc.column
		}
		*issues = append(*issues, issue)
	}
	for _, ch := range u.Errors {
		appendIssuesFromOutput(ch, pointerLC, issues)
	}
}

type lineCol struct {
	line   int
	column int
}

func buildJSONPointerLineMap(raw []byte) map[string]lineCol {
	lineStarts := []int{0}
	for i, b := range raw {
		if b == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	state := pointerWalker{
		dec:        dec,
		raw:        raw,
		prevOffset: 0,
		lineStarts: lineStarts,
		out:        map[string]lineCol{},
	}
	if err := state.parseValue(""); err != nil {
		return map[string]lineCol{}
	}
	return state.out
}

type pointerWalker struct {
	dec        *json.Decoder
	raw        []byte
	prevOffset int
	lineStarts []int
	out        map[string]lineCol
}

func (w *pointerWalker) parseValue(pointer string) error {
	start := skipJSONWhitespace(w.raw, w.prevOffset)
	tok, err := w.dec.Token()
	if err != nil {
		return err
	}
	w.prevOffset = int(w.dec.InputOffset())
	if _, exists := w.out[pointer]; !exists {
		w.out[pointer] = w.offsetToLineCol(start)
	}

	delim, isDelim := tok.(json.Delim)
	if !isDelim {
		return nil
	}
	switch delim {
	case '{':
		for w.dec.More() {
			// key token
			keyTok, err := w.dec.Token()
			if err != nil {
				return err
			}
			w.prevOffset = int(w.dec.InputOffset())
			key, _ := keyTok.(string)
			if err := w.parseValue(joinJSONPointer(pointer, key)); err != nil {
				return err
			}
		}
		_, err := w.dec.Token() // consume '}'
		w.prevOffset = int(w.dec.InputOffset())
		return err
	case '[':
		idx := 0
		for w.dec.More() {
			if err := w.parseValue(fmt.Sprintf("%s/%d", pointer, idx)); err != nil {
				return err
			}
			idx++
		}
		_, err := w.dec.Token() // consume ']'
		w.prevOffset = int(w.dec.InputOffset())
		return err
	default:
		return nil
	}
}

func (w *pointerWalker) offsetToLineCol(offset int) lineCol {
	if offset < 0 {
		offset = 0
	}
	if offset > len(w.raw) {
		offset = len(w.raw)
	}
	i := sort.Search(len(w.lineStarts), func(i int) bool {
		return w.lineStarts[i] > offset
	}) - 1
	if i < 0 {
		i = 0
	}
	lineStart := w.lineStarts[i]
	return lineCol{
		line:   i + 1,
		column: (offset - lineStart) + 1,
	}
}

func skipJSONWhitespace(raw []byte, i int) int {
	for i < len(raw) {
		switch raw[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

func joinJSONPointer(base, seg string) string {
	escaped := strings.ReplaceAll(strings.ReplaceAll(seg, "~", "~0"), "/", "~1")
	return base + "/" + escaped
}

var (
	refPathTailRe   = regexp.MustCompile(` for ([A-Za-z0-9_.\[\]]+)$`)
	invalidJSONRe   = regexp.MustCompile(`^([A-Za-z0-9_.\[\]]+): invalid JSON:`)
	pathTokenExprRe = regexp.MustCompile(`([A-Za-z0-9_]+)|\[(\d+)\]`)
)

func referenceValidationIssue(raw []byte, err error) ValidationIssue {
	msg := err.Error()
	issue := ValidationIssue{
		Phase:   "reference",
		Message: msg,
	}
	pathExpr := ""
	if m := refPathTailRe.FindStringSubmatch(msg); len(m) == 2 {
		pathExpr = m[1]
	} else if m := invalidJSONRe.FindStringSubmatch(msg); len(m) == 2 {
		pathExpr = m[1]
	}
	if pathExpr == "" {
		return issue
	}

	ptr := pathExprToPointer(pathExpr)
	issue.InstancePath = ptr
	if ptr == "" {
		return issue
	}

	if lc, ok := buildJSONPointerLineMap(raw)[ptr]; ok {
		issue.Line = lc.line
		issue.Column = lc.column
	}
	return issue
}

func pathExprToPointer(pathExpr string) string {
	if pathExpr == "" {
		return ""
	}
	matches := pathTokenExprRe.FindAllStringSubmatch(pathExpr, -1)
	if len(matches) == 0 {
		return ""
	}
	segs := make([]string, 0, len(matches))
	for _, m := range matches {
		if m[1] != "" {
			segs = append(segs, m[1])
		} else if m[2] != "" {
			if _, err := strconv.Atoi(m[2]); err == nil {
				segs = append(segs, m[2])
			}
		}
	}
	if len(segs) == 0 {
		return ""
	}
	return "/" + strings.Join(segs, "/")
}
