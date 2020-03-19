package storage

import (
	"bytes"

	"github.com/jmhodges/levigo"
)

type PrefixIterator struct {
	ro         *levigo.ReadOptions
	it         *levigo.Iterator
	prefix     []byte
	shouldStop bool
}

func NewPrefixIterator(db *levigo.DB, prefix []byte) *PrefixIterator {
	ro := levigo.NewReadOptions()
	it := db.NewIterator(ro)
	if len(prefix) > 0 {
		it.Seek(prefix)
	} else {
		it.SeekToFirst()
	}
	return &PrefixIterator{ro, it, prefix, false}
}

func (pi *PrefixIterator) valid() bool {
	return !pi.shouldStop && pi.it.Valid() && bytes.HasPrefix(pi.it.Key(), pi.prefix)
}

func (pi *PrefixIterator) next() {
	pi.it.Next()
}

func (pi *PrefixIterator) close() {
	pi.it.Close()
	pi.ro.Close()
}

func (pi *PrefixIterator) Key() []byte {
	return pi.it.Key()
}

func (pi *PrefixIterator) Value() []byte {
	return pi.it.Value()
}

func (pi *PrefixIterator) Stop() {
	pi.shouldStop = true
}
