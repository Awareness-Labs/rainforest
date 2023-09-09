package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// 1. 定義 Writer 介面
type Writer interface {
	Write(dataStr string) error
}

type JSONFileWriterFactory struct{}

func (f *JSONFileWriterFactory) CreateWriter(sync bool, filePath string) Writer {
	if sync {
		return NewSyncJSONFileWriter(filePath)
	}
	return NewAsyncJSONFileWriter(filePath)
}

// 2. 實現 AsyncJSONFileWriter
type AsyncJSONFileWriter struct {
	FilePath string
	Queue    chan map[string]interface{}
}

func NewAsyncJSONFileWriter(filePath string) *AsyncJSONFileWriter {
	j := &AsyncJSONFileWriter{
		FilePath: filePath,
		Queue:    make(chan map[string]interface{}, 4096),
	}

	go j.periodicWrite()
	return j
}

func (j *AsyncJSONFileWriter) Write(dataStr string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err == nil {
		j.Queue <- data
		return nil
	} else {
		return fmt.Errorf("Error parsing JSON: %v", err)
	}
}

func (j *AsyncJSONFileWriter) periodicWrite() {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var currentData []map[string]interface{}
			for len(j.Queue) > 0 {
				currentData = append(currentData, <-j.Queue)
			}
			if len(currentData) > 0 {
				j.appendToFile(currentData)
			}
		}
	}
}

func (j *AsyncJSONFileWriter) appendToFile(data []map[string]interface{}) {
	var existingData []map[string]interface{}
	if fileContent, err := ioutil.ReadFile(j.FilePath); err == nil {
		json.Unmarshal(fileContent, &existingData)
	}
	existingData = append(existingData, data...)
	updatedContent, _ := json.Marshal(existingData)
	ioutil.WriteFile(j.FilePath, updatedContent, os.ModePerm)
}

// 3. 實現 SyncJSONFileWriter
type SyncJSONFileWriter struct {
	FilePath string
}

func NewSyncJSONFileWriter(filePath string) *SyncJSONFileWriter {
	return &SyncJSONFileWriter{FilePath: filePath}
}

func (j *SyncJSONFileWriter) Write(dataStr string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return fmt.Errorf("Error parsing JSON: %v", err)
	}
	return j.appendToFile(data)
}

func (j *SyncJSONFileWriter) appendToFile(data map[string]interface{}) error {
	var existingData []map[string]interface{}
	if fileContent, err := ioutil.ReadFile(j.FilePath); err == nil {
		json.Unmarshal(fileContent, &existingData)
	}
	existingData = append(existingData, data)
	updatedContent, _ := json.Marshal(existingData)
	return ioutil.WriteFile(j.FilePath, updatedContent, os.ModePerm)
}

func main() {
	factory := &JSONFileWriterFactory{}

	writer := factory.CreateWriter(true, "./test_sync.json")
	writer.Write(`{"name": "John"}`)
	writer.Write(`{"name": "Doe"}`)

	writerAsync := factory.CreateWriter(false, "./test_async.json")
	writerAsync.Write(`{"name": "Alice"}`)
	writerAsync.Write(`{"name": "Bob"}`)

	// Sleep to allow async writer to finish
	time.Sleep(2 * time.Second)
}
