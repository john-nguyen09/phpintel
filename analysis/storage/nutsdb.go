package storage

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/xujiajun/nutsdb"
)

type nutsdbDB struct {
	db       *nutsdb.DB
	txMu     *sync.Mutex
	activeTx *nutsdb.Tx
}

var _ DB = (*nutsdbDB)(nil)

var /* const */ DefaultBucket = "phpintel"

func NewNutsDB(path string) (*nutsdbDB, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(path),
		nutsdb.WithRWMode(nutsdb.MMap),
		nutsdb.WithSyncEnable(false),
	)
	if err != nil {
		return nil, err
	}
	return &nutsdbDB{db, &sync.Mutex{}, nil}, nil
}

func (s *nutsdbDB) Close() {
	s.db.Close()
}

func (s *nutsdbDB) Delete(key []byte) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(DefaultBucket, key)
	})
}

func (s *nutsdbDB) Put(key []byte, value []byte) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(DefaultBucket, key, value, nutsdb.Persistent)
	})
}

func (s *nutsdbDB) Get(key []byte) ([]byte, error) {
	results := []byte(nil)
	err := s.db.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(DefaultBucket, key)
		if err != nil {
			return err
		}
		results = make([]byte, len(entry.Value))
		copy(results, entry.Value)
		return nil
	})
	if err != nil {
		return results, err
	}
	return results, err
}

func (s *nutsdbDB) WriteBatch(f func(Batch) error) error {
	b := newNutsdbDBBatch(s.db)
	err := f(b)
	if err == nil {
		err = s.Commit(b)
	}
	return err
}

func (s *nutsdbDB) PrefixStream(prefix []byte, onData func(Iterator)) {
	var it *nutsdbDBPrefixIterator
	if s.activeTx != nil {
		it = newNutsdbDBPrefixIterator(s.activeTx, prefix)
	} else {
		s.db.View(func(tx *nutsdb.Tx) error {
			it = newNutsdbDBPrefixIterator(tx, prefix)
			return nil
		})
	}
	defer it.close()
	for ; it.valid(); it.next() {
		onData(it)
	}
}

func (s *nutsdbDB) Clear() {
	s.txMu.Lock()
	defer s.txMu.Unlock()
	if err := s.WriteBatch(func(b Batch) error {
		s.activeTx = b.(*nutsdbDBBatch).tx
		s.PrefixStream(nil, func(it Iterator) {
			b.Delete(it.Key())
		})
		return nil
	}); err != nil {
		log.Printf("nutsdbDB.Clear error: %v", err)
	}
	s.activeTx = nil
}

func (s *nutsdbDB) Commit(b *nutsdbDBBatch) error {
	return b.Commit()
}

type nutsdbDBPrefixIterator struct {
	entries    nutsdb.Entries
	index      int64
	shouldStop bool
}

var _ Iterator = (*nutsdbDBPrefixIterator)(nil)

func newNutsdbDBPrefixIterator(tx *nutsdb.Tx, prefix []byte) *nutsdbDBPrefixIterator {
	readEntries := nutsdb.Entries{}
	if len(prefix) == 0 {
		var err error
		readEntries, err = tx.GetAll(DefaultBucket)

		if err != nil {
			log.Printf("newNutsdbDBPrefixIterator error: %v", err)
		}
	}
	if entries, _, err := tx.PrefixScan(DefaultBucket, prefix, 0, nutsdb.ScanNoLimit); err != nil {
		if err != nil {
			log.Printf("newNutsdbDBPrefixIterator error: %v", err)
		}
	} else {
		readEntries = entries
	}
	return &nutsdbDBPrefixIterator{readEntries, 0, false}
}

func (pi *nutsdbDBPrefixIterator) valid() bool {
	length := int64(len(pi.entries))
	return !pi.shouldStop && length > 0 && pi.index < length
}

func (pi *nutsdbDBPrefixIterator) next() {
	if pi.shouldStop {
		return
	}
	atomic.AddInt64(&pi.index, 1)
	pi.index++
}

func (pi *nutsdbDBPrefixIterator) close() {
}

func (pi *nutsdbDBPrefixIterator) Key() []byte {
	return pi.entries[pi.index].Key
}

func (pi *nutsdbDBPrefixIterator) Value() []byte {
	return pi.entries[pi.index].Value
}

func (pi *nutsdbDBPrefixIterator) Stop() {
	pi.shouldStop = true
}

type nutsdbDBBatch struct {
	tx *nutsdb.Tx
}

var _ Batch = (*nutsdbDBBatch)(nil)

func newNutsdbDBBatch(db *nutsdb.DB) *nutsdbDBBatch {
	tx, err := db.Begin(true)
	if err != nil {
		log.Fatalf("newNutsdbDBBatch error: %v", err)
		tx = nil
	}
	return &nutsdbDBBatch{
		tx: tx,
	}
}

func (b *nutsdbDBBatch) Delete(key []byte) {
	if b.tx == nil {
		return
	}
	b.tx.Delete(DefaultBucket, key)
}

func (b *nutsdbDBBatch) Put(key []byte, value []byte) {
	if b.tx == nil {
		return
	}
	b.tx.Put(DefaultBucket, key, value, nutsdb.Persistent)
}

func (b *nutsdbDBBatch) Commit() error {
	return b.tx.Commit()
}
