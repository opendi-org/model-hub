package database

// digest.go — content-addressing for CDM documents.
//
// Responsibilities:
//   1. Canonicalize raw JSON blobs (stable byte representation for hashing).
//   2. Compute content-addressed UUIDs for every CDM child type and the root.
//   3. Rewrite cross-references inside JSONB blobs after child UUIDs are known.
//   4. Strip hub-managed meta fields so they never affect the content hash.
//
// All functions are pure (no I/O, no global state).
// The single entry point used by store.go is assignDigests.

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// ── Hub-managed field clearing ────────────────────────────────────────────────

// clearHubMeta zeroes the fields that the hub manages and that must never
// influence a content hash: creator, updator, createdDate, updatedDate.
// meta.uuid is handled separately (zeroed immediately before hashing).
func clearHubMeta(a *cdm.Asset) {
	a.Creator = ""
	a.CreatedDate = ""
	a.Updator = ""
	a.UpdatedDate = ""
}

// ── JSON canonicalization ─────────────────────────────────────────────────────

// canonicalizeJSON parses raw JSON and re-encodes it using Go's encoder, which
// sorts object keys and produces a stable, minimised byte representation.
// If raw is empty or unparseable the original bytes are returned unchanged.
func canonicalizeJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ── UUID derivation ───────────────────────────────────────────────────────────

// sha256UUID takes the SHA256 of b and returns the first 16 bytes formatted as
// a RFC 4122 UUID string with version=5 and variant bits set.
// Collision probability (~2^-61) is negligible for this use case.
func sha256UUID(b []byte) string {
	sum := sha256.Sum256(b)
	u := make([]byte, 16)
	copy(u, sum[:16])
	u[6] = (u[6] & 0x0f) | 0x50 // version 5
	u[8] = (u[8] & 0x3f) | 0x80 // RFC 4122 variant
	h := toHex(u)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func toHex(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = digits[v>>4]
		out[i*2+1] = digits[v&0x0f]
	}
	return string(out)
}

// ── Per-type hash functions ───────────────────────────────────────────────────
// Each function receives a value type (copy) so it can zero meta.uuid without
// mutating the caller's slice element.
// All JSONB fields (Elements, Content, etc.) must be canonicalized before calling.

func hashIOValue(v cdm.IOValue) string {
	v.Meta.UUID = ""
	clearHubMeta(&v.Meta)
	b, _ := json.Marshal(v)
	return sha256UUID(b)
}

func hashEvaluatable(e cdm.EvaluatableAsset) string {
	e.Meta.UUID = ""
	clearHubMeta(&e.Meta)
	b, _ := json.Marshal(e)
	return sha256UUID(b)
}

func hashRunnable(r cdm.RunnableModel) string {
	r.Meta.UUID = ""
	clearHubMeta(&r.Meta)
	b, _ := json.Marshal(r)
	return sha256UUID(b)
}

func hashDiagram(d cdm.Diagram) string {
	d.Meta.UUID = ""
	clearHubMeta(&d.Meta)
	b, _ := json.Marshal(d)
	return sha256UUID(b)
}

func hashControl(c cdm.Control) string {
	c.Meta.UUID = ""
	clearHubMeta(&c.Meta)
	b, _ := json.Marshal(c)
	return sha256UUID(b)
}

// hashModel computes the root CDM digest from:
//   - root meta (uuid and hub-managed fields zeroed)
//   - a sorted, typed list of every child digest
//
// Sorting ensures the hash is independent of insertion order.
func hashModel(m cdm.CausalDecisionModel) string {
	m.Meta.UUID = ""
	clearHubMeta(&m.Meta)

	type childRef struct {
		Type   string `json:"type"`
		Digest string `json:"digest"`
	}

	cap := len(m.RunnableModels) + len(m.Diagrams) +
		len(m.EvaluatableAssets) + len(m.InputOutputValues) + len(m.Controls)
	children := make([]childRef, 0, cap)

	for _, c := range m.RunnableModels {
		children = append(children, childRef{"runnable", c.Meta.UUID})
	}
	for _, c := range m.Diagrams {
		children = append(children, childRef{"diagram", c.Meta.UUID})
	}
	for _, c := range m.EvaluatableAssets {
		children = append(children, childRef{"evaluatable", c.Meta.UUID})
	}
	for _, c := range m.InputOutputValues {
		children = append(children, childRef{"io_value", c.Meta.UUID})
	}
	for _, c := range m.Controls {
		children = append(children, childRef{"control", c.Meta.UUID})
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i].Type != children[j].Type {
			return children[i].Type < children[j].Type
		}
		return children[i].Digest < children[j].Digest
	})

	sig := struct {
		Meta     cdm.Asset  `json:"meta"`
		Children []childRef `json:"children"`
	}{m.Meta, children}

	b, _ := json.Marshal(sig)
	return sha256UUID(b)
}

// ── JSONB reference rewriting ─────────────────────────────────────────────────

// rewriteRunnableElements rewrites evaluatableAsset, inputs[], and outputs[]
// UUIDs inside a RunnableModel's elements[] JSON blob using the provided
// old→new maps. Only entries present in the maps are rewritten; others are left
// untouched. Returns the original bytes unchanged if both maps are empty.
func rewriteRunnableElements(raw json.RawMessage, ioMap, evalMap map[string]string) (json.RawMessage, error) {
	if len(raw) == 0 || (len(ioMap) == 0 && len(evalMap) == 0) {
		return raw, nil
	}
	var elems []map[string]interface{}
	if err := json.Unmarshal(raw, &elems); err != nil {
		return nil, err
	}
	for _, elem := range elems {
		if v, ok := elem["evaluatableAsset"].(string); ok {
			if mapped, ok2 := evalMap[v]; ok2 {
				elem["evaluatableAsset"] = mapped
			}
		}
		rewriteStringSlice(elem, "inputs", ioMap)
		rewriteStringSlice(elem, "outputs", ioMap)
	}
	b, err := json.Marshal(elems)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// rewriteUUIDArray rewrites UUID strings inside a JSON array using m.
// Returns the original bytes unchanged if m is empty.
func rewriteUUIDArray(raw json.RawMessage, m map[string]string) (json.RawMessage, error) {
	if len(raw) == 0 || len(m) == 0 {
		return raw, nil
	}
	var arr []interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, err
	}
	for i, v := range arr {
		if s, ok := v.(string); ok {
			if mapped, ok2 := m[s]; ok2 {
				arr[i] = mapped
			}
		}
	}
	b, err := json.Marshal(arr)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// rewriteStringSlice rewrites a []string field inside an element map in-place.
func rewriteStringSlice(elem map[string]interface{}, field string, m map[string]string) {
	arr, _ := elem[field].([]interface{})
	for i, x := range arr {
		if s, ok := x.(string); ok {
			if mapped, ok2 := m[s]; ok2 {
				arr[i] = mapped
			}
		}
	}
}

// ── Main entry point ──────────────────────────────────────────────────────────

// assignDigests is the single call that store.go makes into this file.
// It canonicalizes all JSONB blobs, computes and assigns content-addressed
// UUIDs to every child, rewrites cross-references to the new UUIDs, then
// computes and assigns the root CDM UUID.
//
// The function mutates m in place and returns the root CDM UUID.
// It must be called after validateSchema and validateRefs, and before any DB writes.
func assignDigests(m *cdm.CausalDecisionModel) (rootUUID string, err error) {

	// ── Pass 1: leaves with no internal references ────────────────────────────
	// IOValues and EvaluatableAssets have no references to other CDM children,
	// so they can be hashed first. We record old→new UUID maps so that anything
	// referencing them can be updated in the passes below.

	ioOldToNew := make(map[string]string, len(m.InputOutputValues))
	for i := range m.InputOutputValues {
		v := &m.InputOutputValues[i]
		if v.Data, err = canonicalizeJSON(v.Data); err != nil {
			return "", fmt.Errorf("inputOutputValues[%d].data: %w", i, err)
		}
		old := v.Meta.UUID
		v.Meta.UUID = hashIOValue(*v)
		clearHubMeta(&v.Meta)
		if old != "" && old != v.Meta.UUID {
			ioOldToNew[old] = v.Meta.UUID
		}
	}

	evalOldToNew := make(map[string]string, len(m.EvaluatableAssets))
	for i := range m.EvaluatableAssets {
		e := &m.EvaluatableAssets[i]
		if e.Content, err = canonicalizeJSON(e.Content); err != nil {
			return "", fmt.Errorf("evaluatableAssets[%d].content: %w", i, err)
		}
		old := e.Meta.UUID
		e.Meta.UUID = hashEvaluatable(*e)
		clearHubMeta(&e.Meta)
		if old != "" && old != e.Meta.UUID {
			evalOldToNew[old] = e.Meta.UUID
		}
	}

	// ── Pass 2: Diagrams (no cross-references to other children) ─────────────

	for i := range m.Diagrams {
		d := &m.Diagrams[i]
		if d.Elements, err = canonicalizeJSON(d.Elements); err != nil {
			return "", fmt.Errorf("diagrams[%d].elements: %w", i, err)
		}
		if d.Dependencies, err = canonicalizeJSON(d.Dependencies); err != nil {
			return "", fmt.Errorf("diagrams[%d].dependencies: %w", i, err)
		}
		if d.Addons, err = canonicalizeJSON(d.Addons); err != nil {
			return "", fmt.Errorf("diagrams[%d].addons: %w", i, err)
		}
		d.Meta.UUID = hashDiagram(*d)
		clearHubMeta(&d.Meta)
	}

	// ── Pass 3: RunnableModels (reference IOValues and EvaluatableAssets) ─────

	for i := range m.RunnableModels {
		r := &m.RunnableModels[i]
		if r.Elements, err = rewriteRunnableElements(r.Elements, ioOldToNew, evalOldToNew); err != nil {
			return "", fmt.Errorf("runnableModels[%d].elements rewrite: %w", i, err)
		}
		if r.Elements, err = canonicalizeJSON(r.Elements); err != nil {
			return "", fmt.Errorf("runnableModels[%d].elements: %w", i, err)
		}
		if r.Addons, err = canonicalizeJSON(r.Addons); err != nil {
			return "", fmt.Errorf("runnableModels[%d].addons: %w", i, err)
		}
		r.Meta.UUID = hashRunnable(*r)
		clearHubMeta(&r.Meta)
	}

	// ── Pass 4: Controls (reference IOValues and Diagram displays) ────────────

	for i := range m.Controls {
		c := &m.Controls[i]
		if c.InputOutputValues, err = rewriteUUIDArray(c.InputOutputValues, ioOldToNew); err != nil {
			return "", fmt.Errorf("controls[%d].inputOutputValues rewrite: %w", i, err)
		}
		if c.InputOutputValues, err = canonicalizeJSON(c.InputOutputValues); err != nil {
			return "", fmt.Errorf("controls[%d].inputOutputValues: %w", i, err)
		}
		if c.Displays, err = canonicalizeJSON(c.Displays); err != nil {
			return "", fmt.Errorf("controls[%d].displays: %w", i, err)
		}
		c.Meta.UUID = hashControl(*c)
		clearHubMeta(&c.Meta)
	}

	// ── Pass 5: Root addons, then root CDM hash ───────────────────────────────

	if m.Addons, err = canonicalizeJSON(m.Addons); err != nil {
		return "", fmt.Errorf("root addons: %w", err)
	}

	rootUUID = hashModel(*m)
	m.Meta.UUID = rootUUID
	clearHubMeta(&m.Meta)

	return rootUUID, nil
}
