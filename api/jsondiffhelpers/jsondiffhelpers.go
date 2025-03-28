//
// COPYRIGHT OpenDI
//

//hacky way to get some jsondiff functionality to be public (apply method) and made my own jsondiff functionality (invert method) because the library didn't have it at the time of writing. curious...

package jsondiffhelpers

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	jsondiff "github.com/wI2L/jsondiff"
)

//wrote this file because the jsondiff library didn't have an invert patch function at the time of writing. curious...

// JSON Patch operation types.
// These are defined in RFC 6902 section 4.
// https://datatracker.ietf.org/doc/html/rfc6902#section-4
const (
	OperationAdd     = "add"
	OperationReplace = "replace"
	OperationRemove  = "remove"
	OperationMove    = "move"
	OperationCopy    = "copy"
	OperationTest    = "test"
)

// inverts the given patch and applies it to the current model bytes
func ApplyInvertedPatch(currModelBytes []byte, patchBytes []byte) ([]byte, error) {
	// Create a new patch object
	var patch jsondiff.Patch
	//convert from byte array to patch object
	err := json.Unmarshal(patchBytes, &patch)
	if err != nil {

		return nil, fmt.Errorf("error unmarshalling patch: %v", err)
	}

	invertedPatch, _ := InvertPatch(patch)
	//apply the inverted patch to the current JSON bytes we have

	//get byte array form of JSON form of inverted ptach
	// Marshal the struct into JSON
	invertedPatchBytes, err := json.Marshal(invertedPatch)
	if err != nil {
		return nil, fmt.Errorf("error marshalling inverted patch: %v", err)
	}
	jsonpatchPatch, err := jsonpatch.DecodePatch(invertedPatchBytes)
	if err != nil {
		return nil, fmt.Errorf("error decoding inverted patch: %v", err)
	}

	//apply the patch
	modified, err := jsonpatchPatch.Apply(currModelBytes)
	return modified, err
}

// InvertPatch inverts a JSON Patch, preparing the patch to reverse the operations.
// It supports "add", "remove", and "replace" operations for now.
// The returned patch is still invertible.
func InvertPatch(patch jsondiff.Patch) (jsondiff.Patch, error) {
	var invertedPatch jsondiff.Patch

	var prevTestOp *jsondiff.Operation

	for _, op := range patch {
		switch op.Type {
		case OperationAdd:
			// Add operation is inverted by a remove operation with the same path
			//to make it invertible, we first add a test operation with the value now removed.
			invertedPatch = append(invertedPatch, jsondiff.Operation{
				Type:  OperationTest,
				Path:  op.Path,
				Value: op.Value,
			})
			//then we add the remove operation
			invertedPatch = append(invertedPatch, jsondiff.Operation{
				Type: OperationRemove,
				Path: op.Path,
			})
		case OperationRemove:
			// Remove operation is inverted by an add operation with the same path. The value is taken from the previous test operation.
			if prevTestOp == nil {
				return nil, fmt.Errorf("missing test operation for remove operation")
			}
			invertedPatch = append(invertedPatch, jsondiff.Operation{
				Type:  OperationAdd,
				Path:  op.Path,
				Value: prevTestOp.Value,
			})
		case OperationReplace:
			// Replace operation is inverted by:
			// - a test operation to hold the now-previous value of the replace
			invertedPatch = append(invertedPatch, jsondiff.Operation{
				Type:  OperationTest,
				Path:  op.Path,
				Value: op.Value,
			})
			// - the new replace operation replacing it with the now-new value.
			invertedPatch = append(invertedPatch, jsondiff.Operation{
				Type:  OperationReplace,
				Path:  op.Path,
				Value: prevTestOp.Value,
			})
		case OperationTest:
			// store the test operation as it holds the previous value of a remove or replace
			prevTestOp = &op
		default:
			return nil, fmt.Errorf("unsupported operation: %s", op.Type)
		}
	}

	return invertedPatch, nil
}
