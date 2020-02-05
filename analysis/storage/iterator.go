package storage

import (
	"bytes"

	"github.com/kezhuw/leveldb"
)

type PrefixIterator struct {
	it         leveldb.Iterator
	prefix     []byte
	shouldStop bool
}

func NewPrefixIterator(db *leveldb.DB, prefix []byte) *PrefixIterator {
	it := db.Prefix(prefix, nil)
	return &PrefixIterator{it, prefix, false}
}

func (pi *PrefixIterator) valid() bool {
	if pi.shouldStop {
		return false
	}
	return pi.it.Valid() && bytes.HasPrefix(pi.it.Key(), pi.prefix)
}

func (pi *PrefixIterator) next() {
	pi.it.Next()
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
