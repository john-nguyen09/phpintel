package analysis

import (
	"log"

	"github.com/dgraph-io/badger"
)

const collectionSep string = "#"

type entry struct {
	key        []byte
	serialiser *Serialiser
}

func newEntry(prefix string, key string) *entry {
	return &entry{
		key:        []byte(prefix + collectionSep + key),
		serialiser: NewSerialiser(),
	}
}

func (s *entry) getSerialiser() *Serialiser {
	return s.serialiser
}

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) getBytes() []byte {
	return s.serialiser.GetBytes()
}

func writeEntry(txn *badger.Txn, entry *entry) error {
	return txn.Set(entry.getKeyBytes(), entry.getBytes())
}

func writeDocument(db *badger.DB, document *Document) {
	txn := db.NewTransaction(true)
	for _, child := range document.Children {
		if serialisable, ok := child.(Serialisable); ok {
			entry := newEntry(serialisable.GetCollection(), serialisable.GetKey())
			serialisable.Serialise(entry.serialiser)
			if err := writeEntry(txn, entry); err == badger.ErrTxnTooBig {
				txn.Commit()
				txn = db.NewTransaction(true)
				writeEntry(txn, entry)
			}

		}
	}
	err := txn.Commit()
	if err != nil {
		log.Print(err)
	}
}

func getClass(db *badger.DB, name string) []*Class {
	prefix := []byte("class" + collectionSep + name + KeySep)
	classes := []*Class{}
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				classes = append(classes, ReadClass(serialiser))
				return nil
			})
		}
		return nil
	})
	return classes
}
