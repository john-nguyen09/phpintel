package storage

import (
	"github.com/jmhodges/levigo"
)

type Batch struct {
	wb *levigo.WriteBatch
}

func NewBatch() Batch {
	return Batch{
		wb: levigo.NewWriteBatch(),
	}
}

func finaliseBatch(b *Batch) {
	b.wb.Close()
}

func (b *Batch) Delete(key []byte) {
	b.wb.Delete(key)
}

func (b *Batch) Put(key []byte, value []byte) {
	b.wb.Put(key, value)
}

func (b *Batch) Write(s *Storage) error {
	wo := levigo.NewWriteOptions()
	defer wo.Close()
	defer b.wb.Close()
	return s.db.Write(wo, b.wb)
}
