package storage

import (
	"github.com/kezhuw/leveldb"
)

type PrefixIterator struct {
	it         leveldb.Iterator
	shouldStop bool
}

func NewPrefixIterator(db *leveldb.DB, prefix []byte) *PrefixIterator {
	it := db.Prefix(prefix, nil)
	return &PrefixIterator{it, false}
}

func (pi *PrefixIterator) next() bool {
	return pi.it.Next()
}

func (pi *PrefixIterator) close() {
	pi.it.Close()
}

func (pi *PrefixIterator) Key() []byte {
	key := pi.it.Key()
	return append(key[:0:0], key...)
}

func (pi *PrefixIterator) Value() []byte {
	value := pi.it.Value()
	return append(value[:0:0], value...)
}

func (pi *PrefixIterator) Stop() {
	pi.shouldStop = true
}
