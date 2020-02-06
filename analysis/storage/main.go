package storage

import (
	"github.com/kezhuw/leveldb"
)

type Storage struct {
	db *leveldb.DB
}

type StreamOptions struct {
	OnData func(key []byte, val []byte) bool
	OnEnd  func()
}

func NewStorage(path string) (*Storage, error) {
	opts := &leveldb.Options{
		CreateIfMissing: true,
	}
	db, err := leveldb.Open(path, opts)
	if err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) Delete(key []byte) error {
	return s.db.Delete(key, nil)
}

func (s *Storage) Put(key []byte, value []byte) error {
	return s.db.Put(key, value, nil)
}

func (s *Storage) Get(key []byte) ([]byte, error) {
	return s.db.Get(key, nil)
}

func (s *Storage) WriteBatch(f func(*leveldb.Batch) error) error {
	var wb leveldb.Batch
	err := f(&wb)
	if err == nil {
		err = s.db.Write(wb, nil)
	}
	return err
}

func (s *Storage) PrefixStream(prefix []byte, onData func(*PrefixIterator)) {
	it := NewPrefixIterator(s.db, prefix)
	defer it.close()
	for it.next() {
		onData(it)
	}
}

func (s *Storage) Clear() {
	s.WriteBatch(func(wb *leveldb.Batch) error {
		s.PrefixStream(nil, func(it *PrefixIterator) {
			wb.Delete(it.Key())
		})
		return nil
	})
}
