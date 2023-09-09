package storage

import (
	apiv1 "github.com/Awareness-Labs/rainforest/pkg/proto/api/v1"
)

type Storage interface {
	AtomicStorage
	SinkStorage
}

type AtomicStorage interface {
	Scan(start []byte, end []byte, reverse bool, limit int) ([]*apiv1.KeyValue, error)
	Set(key, val []byte) error
	GC()
}

type SinkStorage interface {
}

type storage struct {
	AtomicStorage
	SinkStorage
}

func CombineStorageEngine(atomic AtomicStorage, sink SinkStorage) Storage {
	return &storage{atomic, sink}
}
