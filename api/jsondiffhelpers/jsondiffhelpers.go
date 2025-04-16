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
	//https://github.com/evanphx/json-patch?tab=BSD-3-Clause-1-ov-file
	// see BSD-3 License for licensing details.
	"github.com/qri-io/jsonpointer"
	jsondiff "github.com/wI2L/jsondiff"
)

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

	invertedPatch, _, _ := InvertPatch(patch, currModelBytes)
	//apply the inverted patch to the current JSON bytes we have

	//get byte array form of JSON form of inverted ptach
	// Marshal the struct into JSON
	invertedPatchBytes, err := json.Marshal(invertedPatch)
	if err != nil {
		return nil, fmt.Errorf("error marshalling inverted patch: %v", err)
	}
	//fmt.Println("Inverted Patch is:")
	//fmt.Println(string(invertedPatchBytes))

	//we need to use the jsonpatch library to actually apply the patch. Given that it's the same struct structurally, the byte form will decode properly .
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

// TODO there is a bug with this. fix it!
// InvertPatch inverts a JSON Patch, preparing the patch to reverse the operations.
// It supports "add", "remove", and "replace" operations for now.
// The returned patch is still invertible.
// This also returns the results of applying the inverted patch on the original json.
func InvertPatch(patch jsondiff.Patch, originalJSON []byte) (jsondiff.Patch, []byte, error) {
	//TODO the creator of the jsondiff library should have added an invert method by the time a new team gets this code. Then, we have no need for this method.
	var invertedPatch jsondiff.Patch

	// Create a new slice with the same length
	//change tracker tracks the changes made to the original JSON.
	/*
		So, let's say we have an array.
		{
							"array": [1, 2, 3]
		}
				We have a patch that we want to invert.
				The patch is constituted of:
				op = add, path = /array/-, val = 4
				op = add, path = /array/-, val = 5

				This would result in the array [1, 2, 3, 4, 5] obviously.
				But what does it look like when we invert it?

				[1, 2, 3, 4, 5] -> [1, 2, 3]
				op = reomve, path = /array/4, val = 5
				op = remove, path = /array/3, val = 4

				//so, we need to approach the patch's operations in reverse order. while doing so, we need to keep track of changes this inverted patch would make to the JSON.








	*/
	changeTracker := make([]byte, len(originalJSON))

	// Copy the data
	copy(changeTracker, originalJSON)

	var prevTestOp *jsondiff.Operation
	//we need to iterate through patch in reverse to properly invert it.
	//for each operation in the patch, we need to invert it.
	for i := len(patch) - 1; i >= 0; i-- {
		op := patch[i]
		if i > 0 {
			prevTestOp = &patch[i-1] //we get the corresponding test operation if the operation is an remove or replace operation.

		}
		//fmt.Println("Path: ", op.Path)
		switch op.Type {
		case OperationAdd:

			//add gets a bit complicated
			invertedPatch2 := invertOperationAdd(invertedPatch, op, changeTracker)
			invertedPatch = invertedPatch2

		case OperationRemove:
			// Remove operation is inverted by an add operation with the same path. The value is taken from the previous test operation.
			if prevTestOp == nil {
				return nil, nil, fmt.Errorf("missing test operation for remove operation")
			}
			invertedPatch = append(invertedPatch, invertOperationRemove(op, changeTracker, *prevTestOp))

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
			// do nothing
		default:
			return nil, nil, fmt.Errorf("unsupported operation: %s", op.Type)
		}

		//apply the inverted operation to the changetracker.
		if op.Type != OperationTest {
			//get what we just appended to the inverted patch
			invertedOp := invertedPatch[len(invertedPatch)-1]
			var tempPatch jsondiff.Patch
			tempPatch = append(tempPatch, invertedOp)
			//apply the patch to the change tracker.
			tempPatchBytes, _ := json.Marshal(tempPatch)

			//we need to use the jsonpatch library to actually apply the patch. Given that it's the same struct structurally, the byte form will decode properly .
			jsonpatchPatch, _ := jsonpatch.DecodePatch(tempPatchBytes)
			//apply the patch
			modified, _ := jsonpatchPatch.Apply(changeTracker)

			//update the change tracker with the modified JSON.
			changeTracker = modified

		}
	}

	return invertedPatch, changeTracker, nil
}

func getAllCharactersAfterLastSlash(path string) string {
	// Find the last occurrence of "/"
	lastSlashIndex := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastSlashIndex = i
			break
		}
	}

	// If no "/" is found, return the entire string
	if lastSlashIndex == -1 {
		return path
	}

	// Return the substring after the last "/"
	return path[lastSlashIndex+1:]
}

func getAllCharactersBeforeLastSlash(path string) string {
	// Find the last occurrence of "/"
	lastSlashIndex := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastSlashIndex = i
			break
		}
	}

	// If no "/" is found, return the entire string
	if lastSlashIndex == -1 {
		return path
	}

	// Return the substring before the last "/"
	return path[:lastSlashIndex]
}

func invertOperationRemove(op jsondiff.Operation, changeTracker []byte, prevTestOp jsondiff.Operation) jsondiff.Operation {

	//check if the remove is removing the last index of the array. If so, we need an add operation that adds to the end of the array using the "-" index.
	finalPathElement := getAllCharactersAfterLastSlash(op.Path)
	idx, err := strconv.Atoi(finalPathElement)
	if err == nil {
		//this must mean the path is to an array.
		newPath := getAllCharactersBeforeLastSlash(op.Path)
		array, _ := GetJSONByPath(changeTracker, newPath)
		length, _, _ := getLastElement(array)
		if idx == length-1 {
			//we've confirmed that this remove is removing the last index of the array.
			return jsondiff.Operation{
				Type:  OperationAdd,
				Path:  newPath + "/-",
				Value: prevTestOp.Value,
			}
		}
	}

	//otherwise just return an add operation with the same value and path.

	return jsondiff.Operation{
		Type:  OperationAdd,
		Path:  op.Path,
		Value: prevTestOp.Value,
	}
}

func invertOperationAdd(invertedPatch jsondiff.Patch, op jsondiff.Operation, changeTracker []byte) jsondiff.Patch {

	//if the add has a - at the end, it is appending to the end of the array at the path.
	if op.Path[len(op.Path)-1] == '-' {
		//remove the - from the path as ewll as the trailing slash
		newPath := op.Path[:len(op.Path)-2]
		array, _ := GetJSONByPath(changeTracker, newPath)
		length, _, _ := getLastElement(array)

		//toPrint := fmt.Sprintf("%T", op.Value)
		//fmt.Println("Type of v is: ", toPrint)

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
	parsed := map[string]interface{}{}
	json.Unmarshal([]byte(jsonText), &parsed)

	//create a JSON pointer
	pointer, _ := jsonpointer.Parse(string(path))
	value, _ := pointer.Eval(parsed)
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
