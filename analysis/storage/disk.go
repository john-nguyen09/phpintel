package storage

import (
	"bytes"

	"github.com/jmhodges/levigo"
)

type disk struct {
	db *levigo.DB
}

var _ DB = (*disk)(nil)

func NewDisk(path string) (*disk, error) {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3 << 30))
	opts.SetCreateIfMissing(true)
	opts.SetCompression(levigo.SnappyCompression)
	db, err := levigo.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &disk{db}, nil
}

func (s *disk) Close() {
	s.db.Close()
}

func (s *disk) Delete(key []byte) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	return s.db.Delete(wo, key)
}

func (s *disk) Put(key []byte, value []byte) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	return s.db.Put(wo, key, value)
}

func (s *disk) Get(key []byte) ([]byte, error) {
	ro := levigo.NewReadOptions()
	defer ro.Close()
	return s.db.Get(ro, key)
}

func (s *disk) WriteBatch(f func(Batch) error) error {
	b := newDiskBatch()
	err := f(b)
	if err == nil {
		err = s.Commit(b)
	}
	return err
}

func (s *disk) PrefixStream(prefix []byte, onData func(Iterator)) {
	it := newDiskPrefixIterator(s.db, prefix)
	defer it.close()
	for ; it.valid(); it.next() {
		onData(it)
	}
}

func (s *disk) Clear() {
	s.WriteBatch(func(b Batch) error {
		s.PrefixStream(nil, func(it Iterator) {
			b.Delete(it.Key())
		})
		return nil
	})
}

func (s *disk) Commit(b *diskBatch) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	defer b.wb.Close()
	return s.db.Write(wo, b.wb)
}

type diskPrefixIterator struct {
	ro         *levigo.ReadOptions
	it         *levigo.Iterator
	prefix     []byte
	shouldStop bool
}

var _ Iterator = (*diskPrefixIterator)(nil)

func newDiskPrefixIterator(db *levigo.DB, prefix []byte) *diskPrefixIterator {
	ro := levigo.NewReadOptions()
	it := db.NewIterator(ro)
	if len(prefix) > 0 {
		it.Seek(prefix)
	} else {
		it.SeekToFirst()
	}
	return &diskPrefixIterator{ro, it, prefix, false}
}

func (pi *diskPrefixIterator) valid() bool {
	return !pi.shouldStop && pi.it.Valid() && bytes.HasPrefix(pi.it.Key(), pi.prefix)
}

func (pi *diskPrefixIterator) next() {
	if pi.shouldStop {
		return
	}
	pi.it.Next()
}

func (pi *diskPrefixIterator) close() {
	pi.it.Close()
	pi.ro.Close()
}

func (pi *diskPrefixIterator) Key() []byte {
	return pi.it.Key()
}

func (pi *diskPrefixIterator) Value() []byte {
	return pi.it.Value()
}

func (pi *diskPrefixIterator) Stop() {
	pi.shouldStop = true
}

type diskBatch struct {
	wb *levigo.WriteBatch
}

var _ Batch = (*diskBatch)(nil)

func newDiskBatch() *diskBatch {
	return &diskBatch{
		wb: levigo.NewWriteBatch(),
	}
}

func (b *diskBatch) Delete(key []byte) {
	b.wb.Delete(key)
}

func (b *diskBatch) Put(key []byte, value []byte) {
	b.wb.Put(key, value)
}
