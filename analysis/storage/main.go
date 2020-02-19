package storage

import (
	"github.com/jmhodges/levigo"
)

type Storage struct {
	db *levigo.DB
}

type StreamOptions struct {
	OnData func(key []byte, val []byte) bool
	OnEnd  func()
}

func NewStorage(path string) (*Storage, error) {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3 << 30))
	opts.SetCreateIfMissing(true)
	db, err := levigo.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) Delete(key []byte) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	return s.db.Delete(wo, key)
}

func (s *Storage) Put(key []byte, value []byte) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	return s.db.Put(wo, key, value)
}

func (s *Storage) Get(key []byte) ([]byte, error) {
	ro := levigo.NewReadOptions()
	defer ro.Close()
	return s.db.Get(ro, key)
}

func (s *Storage) WriteBatch(f func(*Batch) error) error {
	b := NewBatch()
	err := f(&b)
	if err == nil {
		err = b.Write(s)
	}
	return err
}

func (s *Storage) PrefixStream(prefix []byte, onData func(*PrefixIterator)) {
	it := NewPrefixIterator(s.db, prefix)
	defer it.close()
	for ; it.valid(); it.next() {
		onData(it)
	}
}

func (s *Storage) Clear() {
	s.WriteBatch(func(b *Batch) error {
		s.PrefixStream(nil, func(it *PrefixIterator) {
			b.Delete(it.Key())
		})
		return nil
	})
}
