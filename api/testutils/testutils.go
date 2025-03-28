package testutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Function to read and parse JSON file into given object
func LoadJSONFromFile(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(byteValue, v)
}
