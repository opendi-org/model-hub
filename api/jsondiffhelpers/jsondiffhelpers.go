//
// COPYRIGHT OpenDI
//

//hacky way to get some jsondiff functionality to be public (apply method) and made my own jsondiff functionality (invert method) because the library didn't have it at the time of writing. curious...

package jsondiffhelpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	jsonpatch "github.com/evanphx/json-patch"
	jsondiff "github.com/wI2L/jsondiff"
	"github.com/xeipuuv/gojsonpointer"
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
	//fmt.Println("Patch is:")
	//fmt.Println((string(patchBytes)))

	invertedPatch, _ := InvertPatch(patch, currModelBytes)
	//apply the inverted patch to the current JSON bytes we have

	//get byte array form of JSON form of inverted ptach
	// Marshal the struct into JSON
	invertedPatchBytes, err := json.Marshal(invertedPatch)
	if err != nil {
		return nil, fmt.Errorf("error marshalling inverted patch: %v", err)
	}
	//fmt.Println("Inverted Patch is:")
	//fmt.Println(string(invertedPatchBytes))
	jsonpatchPatch, err := jsonpatch.DecodePatch(invertedPatchBytes)
	if err != nil {
		return nil, fmt.Errorf("error decoding inverted patch: %v", err)
	}

	//apply the patch
	modified, err := jsonpatchPatch.Apply(currModelBytes)
	if err != nil {
		fmt.Println("Error is in .Apply()")
		fmt.Println(err.Error())
	}
	return modified, err
}

// InvertPatch inverts a JSON Patch, preparing the patch to reverse the operations.
// It supports "add", "remove", and "replace" operations for now.
// The returned patch is still invertible.
func InvertPatch(patch jsondiff.Patch, originalJSON []byte) (jsondiff.Patch, error) {
	//TODO the creator of the jsondiff library should have added an invert method
	var invertedPatch jsondiff.Patch

	var prevTestOp *jsondiff.Operation

	for _, op := range patch {
		//fmt.Println("Path: ", op.Path)
		switch op.Type {
		case OperationAdd:

			//add gets a bit complicated
			invertedPatch2 := invertOperationAdd(invertedPatch, op, originalJSON)
			invertedPatch = invertedPatch2

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

func invertOperationAdd(invertedPatch jsondiff.Patch, op jsondiff.Operation, originalJSON []byte) jsondiff.Patch {

	//if the add has a - at the end, it is appending to the end of the array at the path.
	if op.Path[len(op.Path)-1] == '-' {
		//remove the - from the path as ewll as the trailing slash
		newPath := op.Path[:len(op.Path)-2]
		array, _ := GetJSONByPath(originalJSON, newPath)
		length, _, _ := getLastElement(array)

		toPrint := fmt.Sprintf("%T", op.Value)
		fmt.Println("Type of v is: ", toPrint)

		//add a test operation that checks if the array has the value we got.
		invertedPatch = append(invertedPatch, jsondiff.Operation{
			Type:  OperationTest,
			Path:  newPath + "/" + strconv.Itoa(length-1),
			Value: op.Value,
		})
		//we need the remove operation to remove the length of the array - 1.
		invertedPatch = append(invertedPatch, jsondiff.Operation{
			Type: OperationRemove,
			Path: newPath + "/" + strconv.Itoa(length-1),
		})
		return invertedPatch
	}

	//otherwise, add operation -- whether adding to an array indice or just adding a key to a map, is standard

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
	return invertedPatch
}

// GetJSONByPath returns the JSON of the given path in the JSON document.
func GetJSONByPath(jsonText []byte, path string) ([]byte, error) {
	var jsonDocument map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	//create a JSON pointer
	pointer, _ := gojsonpointer.NewJsonPointer(string(path))
	value, _, _ := pointer.Get(jsonDocument)
	// Marshal the value back to JSON
	jsonValue, _ := json.Marshal(value)
	return jsonValue, nil
}

func getLastElement(jsonArray []byte) (int, []byte, error) {
	// Declare a slice to hold the unmarshalled data
	var array []interface{}

	// Unmarshal the JSON array string into the slice
	if err := json.Unmarshal(jsonArray, &array); err != nil {
		return 0, nil, err
	}
	//get last element
	lastElement, _ := json.Marshal(array[len(array)-1])
	// Return the length of the slice
	return len(array), lastElement, nil
}
