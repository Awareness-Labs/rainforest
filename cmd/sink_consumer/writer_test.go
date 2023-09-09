package main

import (
	"testing"
	"time"
)

func BenchmarkSyncJSONFileWriter(b *testing.B) {
	factory := &JSONFileWriterFactory{}
	writer := factory.CreateWriter(true, "./test_bench_sync.json")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writer.Write(`{"name": "John"}`)
		writer.Write(`{"name": "Doe"}`)
	}
}

func BenchmarkAsyncJSONFileWriter(b *testing.B) {
	factory := &JSONFileWriterFactory{}
	writer := factory.CreateWriter(false, "./test_bench_async.json")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writer.Write(`{"name": "Alice"}`)
		writer.Write(`{"name": "Bob"}`)
	}

	// This is to ensure all the writes are flushed before the benchmark stops
	b.StopTimer()
	time.Sleep(2 * time.Second)
	b.StartTimer()
}
