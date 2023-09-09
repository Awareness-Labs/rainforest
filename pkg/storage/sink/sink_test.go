package sink

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestSinkKeyValue(t *testing.T) {
	dir := "./data"

	// Setup OS based filesystem
	fs := afero.NewOsFs()

	// Overwrite the global file system in jsonStorage
	js := NewJSONStorage(dir).(*jsonStorage)
	js.fs = fs

	// Ensure test directory exists
	if err := fs.MkdirAll(dir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	defer func() {
		_ = fs.RemoveAll(dir) // Uncomment if you want to remove test files and directory after test
	}()

	// Simulated data input for multiple key-values
	filename := "test.json"
	var inputData = make(map[string]string)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		inputData[key] = value
		valueData, _ := json.Marshal(value)
		err := js.SinkJSON(key, valueData)
		if err != nil {
			t.Fatalf("Unexpected error when writing key-value %d: %v", i, err)
		}
	}

	// Sleep a bit to ensure that data has been written due to the buffered channel
	time.Sleep(50 * time.Millisecond)

	// Verify written data
	filePath := dir + "/" + filename
	fileData, err := afero.ReadFile(fs, filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Parse the read data
	var readData map[string]string
	if err := json.Unmarshal(fileData, &readData); err != nil {
		t.Fatalf("Failed to parse read JSON data: %v", err)
	}

	// Compare inputData and readData
	if !reflect.DeepEqual(inputData, readData) {
		t.Errorf("Expected %+v, but got %+v", inputData, readData)
	}
}
