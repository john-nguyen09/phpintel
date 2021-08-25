package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type goLevelDB struct {
	db *leveldb.DB
}

var _ DB = (*goLevelDB)(nil)

func NewGoLevelDB(path string) (*goLevelDB, error) {
	o := &opt.Options{
		Filter: filter.NewBloomFilter(10),
		NoSync: true,
	}
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, err
	}
	return &goLevelDB{db}, nil
}

func (s *goLevelDB) Close() {
	s.db.Close()
}

func (s *goLevelDB) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

func (s *goLevelDB) Put(key []byte, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *goLevelDB) Get(key []byte) ([]byte, error) {
	return s.db.Get(key, nil)
}

func (s *goLevelDB) WriteBatch(f func(Batch) error) error {
	b := newGoLevelDBBatch()
	err := f(b)
	if err == nil {
		err = s.Commit(b)
	}
	return err
}

func (s *goLevelDB) PrefixStream(prefix []byte, onData func(Iterator)) {
	it := newGoLevelDBPrefixIterator(s.db, prefix)
	defer it.close()
	for ; it.valid(); it.next() {
		onData(it)
	}
}

func (s *goLevelDB) Clear() {
	s.WriteBatch(func(b Batch) error {
		s.PrefixStream(nil, func(it Iterator) {
			b.Delete(it.Key())
		})
		return nil
	})
}

func (s *goLevelDB) Commit(b *goLevelDBBatch) error {
	return s.db.Write(b.wb, nil)
}

type goLevelDBPrefixIterator struct {
	it         iterator.Iterator
	prefix     []byte
	shouldStop bool
}

var _ Iterator = (*goLevelDBPrefixIterator)(nil)

func newGoLevelDBPrefixIterator(db *leveldb.DB, prefix []byte) *goLevelDBPrefixIterator {
	var it iterator.Iterator
	if len(prefix) > 0 {
		it = db.NewIterator(util.BytesPrefix(prefix), nil)
	} else {
		it = db.NewIterator(nil, nil)
	}
	it.Next()
	return &goLevelDBPrefixIterator{it, prefix, false}
}

func (pi *goLevelDBPrefixIterator) valid() bool {
	return !pi.shouldStop && pi.it.Valid()
}

func (pi *goLevelDBPrefixIterator) next() {
	if pi.shouldStop {
		return
	}
	pi.it.Next()
}

func (pi *goLevelDBPrefixIterator) close() {
	pi.it.Release()
}

func copyByteSlice(src []byte) []byte {
	result := make([]byte, len(src))
	copy(result, src)
	return result
}

func (pi *goLevelDBPrefixIterator) Key() []byte {
	return copyByteSlice(pi.it.Key())
}

func (pi *goLevelDBPrefixIterator) Value() []byte {
	return copyByteSlice(pi.it.Value())
}

func (pi *goLevelDBPrefixIterator) Stop() {
	pi.shouldStop = true
}

type goLevelDBBatch struct {
	wb *leveldb.Batch
}

var _ Batch = (*goLevelDBBatch)(nil)

func newGoLevelDBBatch() *goLevelDBBatch {
	return &goLevelDBBatch{
		wb: new(leveldb.Batch),
	}
}

func (b *goLevelDBBatch) Delete(key []byte) {
	b.wb.Delete(key)
}

func (b *goLevelDBBatch) Put(key []byte, value []byte) {
	b.wb.Put(key, value)
}
