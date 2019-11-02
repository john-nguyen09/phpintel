package analysis

import (
	"log"
	"strings"
	"sync"

	"github.com/dgraph-io/badger"
)

const collectionSep string = "#"

type entry struct {
	key        []byte
	serialiser *Serialiser
}

func newEntry(collection string, key string) *entry {
	return &entry{
		key:        []byte(collection + collectionSep + key),
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

type Store struct {
	db            *badger.DB
	documentLocks map[string]sync.Mutex
}

func NewStore(storePath string) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(storePath))
	if err != nil {
		return nil, err
	}
	return &Store{
		db:            db,
		documentLocks: map[string]sync.Mutex{},
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SyncDocument(document *Document) {
	db := s.db
	txn := db.NewTransaction(true)
	forgetAllSymbols(db, txn, document)
	writeAllSymbols(db, txn, document)
	err := txn.Commit()
	if err != nil {
		log.Print(err)
	}
}

func forgetAllSymbols(db *badger.DB, txn *badger.Txn, document *Document) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()
	entry := newEntry("documentSymbols", document.GetURI()+KeySep)
	for it.Seek(entry.getKeyBytes()); it.ValidForPrefix(entry.getKeyBytes()); it.Next() {
		item := it.Item()
		keyInfo := strings.Split(string(item.Key()), KeySep)
		toBeDelete := newEntry(keyInfo[1], keyInfo[2])
		if err := deleteEntry(txn, toBeDelete); err == badger.ErrTxnTooBig {
			txn.Commit()
			txn := db.NewTransaction(true)
			deleteEntry(txn, toBeDelete)
		}
	}
}

func writeAllSymbols(db *badger.DB, txn *badger.Txn, document *Document) {
	for _, child := range document.Children {
		if serialisable, ok := child.(Serialisable); ok {
			entry := newEntry(serialisable.GetCollection(), serialisable.GetKey())
			serialisable.Serialise(entry.serialiser)
			if err := writeEntry(txn, entry); err == badger.ErrTxnTooBig {
				txn.Commit()
				txn = db.NewTransaction(true)
				writeEntry(txn, entry)
			}
			if err := rememberSymbol(txn, document, serialisable); err == badger.ErrTxnTooBig {
				txn.Commit()
				txn = db.NewTransaction(true)
				rememberSymbol(txn, document, serialisable)
			}
		}
	}
}

func rememberSymbol(txn *badger.Txn, document *Document, serialisable Serialisable) error {
	entry := newEntry("documentSymbols", document.GetURI()+KeySep+serialisable.GetCollection()+KeySep+serialisable.GetKey())
	return writeEntry(txn, entry)
}

func writeEntry(txn *badger.Txn, entry *entry) error {
	return txn.Set(entry.getKeyBytes(), entry.getBytes())
}

func deleteEntry(txn *badger.Txn, entry *entry) error {
	return txn.Delete(entry.getKeyBytes())
}

func (s *Store) getClasses(name string) []*Class {
	prefix := []byte("class" + collectionSep + name + KeySep)
	classes := []*Class{}
	s.db.View(func(txn *badger.Txn) error {
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

func (s *Store) getInterfaces(name string) []*Interface {
	prefix := []byte("interface" + collectionSep + name + KeySep)
	interfaces := []*Interface{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				interfaces = append(interfaces, ReadInterface(serialiser))
				return nil
			})
		}
		return nil
	})
	return interfaces
}

func (s *Store) getTraits(name string) []*Trait {
	prefix := []byte("trait" + collectionSep + name + KeySep)
	traits := []*Trait{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				traits = append(traits, ReadTrait(serialiser))
				return nil
			})
		}
		return nil
	})
	return traits
}

func (s *Store) getFunctions(name string) []*Function {
	prefix := []byte("function" + collectionSep + name + KeySep)
	functions := []*Function{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				functions = append(functions, ReadFunction(serialiser))
				return nil
			})
		}
		return nil
	})
	return functions
}

func (s *Store) getConsts(name string) []*Const {
	prefix := []byte("const" + collectionSep + name + KeySep)
	consts := []*Const{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				consts = append(consts, ReadConst(serialiser))
				return nil
			})
		}
		return nil
	})
	return consts
}

func (s *Store) getDefines(name string) []*Define {
	prefix := []byte("define" + collectionSep + name + KeySep)
	defines := []*Define{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				defines = append(defines, ReadDefine(serialiser))
				return nil
			})
		}
		return nil
	})
	return defines
}

func (s *Store) getMethods(scope string, name string) []*Method {
	prefix := []byte("method" + collectionSep + scope + KeySep + name + KeySep)
	methods := []*Method{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				methods = append(methods, ReadMethod(serialiser))
				return nil
			})
		}
		return nil
	})
	return methods
}

func (s *Store) getClassConsts(scope string, name string) []*ClassConst {
	prefix := []byte("classConst" + collectionSep + scope + KeySep + name + KeySep)
	classConsts := []*ClassConst{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				classConsts = append(classConsts, ReadClassConst(serialiser))
				return nil
			})
		}
		return nil
	})
	return classConsts
}

func (s *Store) getProperties(scope string, name string) []*Property {
	prefix := []byte("property" + collectionSep + scope + KeySep + name + KeySep)
	properties := []*Property{}
	s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				serialiser := SerialiserFromByteSlice(v)
				properties = append(properties, ReadProperty(serialiser))
				return nil
			})
		}
		return nil
	})
	return properties
}
