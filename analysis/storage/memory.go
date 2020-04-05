package storage

import (
	"fmt"

	iradix "github.com/hashicorp/go-immutable-radix"
)

// Memory is a memory database which uses a radix tree
type memory struct {
	radix *iradix.Tree
}

var _ DB = (*memory)(nil)

func newMemory() *memory {
	return &memory{
		radix: iradix.New(),
	}
}

func (m *memory) Clear() {
	m.radix = iradix.New()
}

func (m *memory) Close() {}

func (m *memory) Delete(key []byte) error {
	m.radix, _, _ = m.radix.Delete(key)
	return nil
}

func (m *memory) Put(key []byte, value []byte) error {
	m.radix, _, _ = m.radix.Insert(key, value)
	return nil
}

func (m *memory) Get(key []byte) ([]byte, error) {
	if value, ok := m.radix.Get(key); ok {
		if bys, ok := value.([]byte); ok {
			return bys, nil
		}
		return nil, fmt.Errorf("%T is not []byte", value)
	}
	return nil, fmt.Errorf("Key not found")
}

func (m *memory) WriteBatch(f func(Batch) error) error {
	b := newMemoryBatch(m)
	err := f(b)
	if err == nil {
		m.radix = b.commit()
	}
	return err
}

func (m *memory) PrefixStream(prefix []byte, onData func(Iterator)) {
	for it := newMemoryIterator(m, prefix); it.valid(); it.next() {
		onData(it)
	}
}

type memoryBatch struct {
	txn *iradix.Txn
}

var _ Batch = (*memoryBatch)(nil)

func newMemoryBatch(m *memory) *memoryBatch {
	return &memoryBatch{
		txn: m.radix.Txn(),
	}
}

func (b *memoryBatch) Delete(key []byte) {
	b.txn.Delete(key)
}

func (b *memoryBatch) Put(key []byte, value []byte) {
	b.txn.Insert(key, value)
}

func (b *memoryBatch) commit() *iradix.Tree {
	return b.txn.Commit()
}

type memoryIterator struct {
	it         *iradix.Iterator
	key        []byte
	value      []byte
	shouldStop bool
	end        bool
}

var _ Iterator = (*memoryIterator)(nil)

func newMemoryIterator(m *memory, prefix []byte) *memoryIterator {
	it := m.radix.Root().Iterator()
	it.SeekPrefix(prefix)

	memIt := &memoryIterator{
		it: it,
	}
	memIt.next()
	return memIt
}

func (it *memoryIterator) Key() []byte {
	return it.key
}

func (it *memoryIterator) Value() []byte {
	return it.value
}

func (it *memoryIterator) Stop() {
	it.shouldStop = true
}

func (it *memoryIterator) close() {
}

func (it *memoryIterator) next() {
	if it.shouldStop {
		return
	}
	if key, value, ok := it.it.Next(); ok {
		it.key = key
		if value, ok := value.([]byte); ok {
			it.value = value
		}
		return
	}
	it.end = true
}

func (it *memoryIterator) valid() bool {
	if it.shouldStop {
		return false
	}
	if it.end {
		return false
	}
	return true
}
