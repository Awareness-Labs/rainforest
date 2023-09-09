package atomic

import (
	"time"

	"github.com/rs/zerolog/log"

	apiv1 "github.com/Awareness-Labs/rainforest/pkg/proto/api/v1"
	"github.com/Awareness-Labs/rainforest/pkg/storage"
	bdb "github.com/dgraph-io/badger/v4"
)

type badger struct {
	db  *bdb.DB
	dir string
}

func NewBadger(dir string) storage.AtomicStorage {
	db, err := bdb.Open(bdb.DefaultOptions(dir))
	if err != nil {
		log.Error().Msg(err.Error())
		return nil
	}
	return &badger{
		db:  db,
		dir: dir,
	}
}

func (b *badger) GC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
		again:
			err := b.db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()
}

func (b *badger) Scan(start []byte, end []byte, reverse bool, limit int) ([]*apiv1.KeyValue, error) {
	kvs := []*apiv1.KeyValue{}
	err := b.db.View(func(txn *bdb.Txn) error {

		opts := bdb.DefaultIteratorOptions
		opts.Reverse = reverse

		it := txn.NewIterator(opts)
		defer it.Close()
		count := 0

		for it.Seek([]byte(start)); it.ValidForPrefix([]byte(end)); it.Next() {
			if count >= limit {
				break
			}
			item := it.Item()
			item.Value(func(v []byte) error {
				kvs = append(kvs, &apiv1.KeyValue{
					Key:   string(item.Key()),
					Value: string(v),
				})
				return nil
			})
			count++
		}
		return nil
	})
	if err != nil {
		return kvs, err
	}
	return kvs, err
}

func (b *badger) Set(key, val []byte) error {
	txn := b.db.NewTransaction(true)
	txn.Set(key, val)
	err := txn.Commit()
	if err != nil {
		log.Error().Msg(err.Error())
		return err
	}
	return nil
}
