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

func getClasses(db *badger.DB, name string) []*Class {
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

func getInterfaces(db *badger.DB, name string) []*Interface {
	prefix := []byte("interface" + collectionSep + name + KeySep)
	interfaces := []*Interface{}
	db.View(func(txn *badger.Txn) error {
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

func getTraits(db *badger.DB, name string) []*Trait {
	prefix := []byte("trait" + collectionSep + name + KeySep)
	traits := []*Trait{}
	db.View(func(txn *badger.Txn) error {
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

func getFunctions(db *badger.DB, name string) []*Function {
	prefix := []byte("function" + collectionSep + name + KeySep)
	functions := []*Function{}
	db.View(func(txn *badger.Txn) error {
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

func getConsts(db *badger.DB, name string) []*Const {
	prefix := []byte("const" + collectionSep + name + KeySep)
	consts := []*Const{}
	db.View(func(txn *badger.Txn) error {
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

func getDefines(db *badger.DB, name string) []*Define {
	prefix := []byte("define" + collectionSep + name + KeySep)
	defines := []*Define{}
	db.View(func(txn *badger.Txn) error {
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

func getMethods(db *badger.DB, scope string, name string) []*Method {
	prefix := []byte("method" + collectionSep + scope + KeySep + name + KeySep)
	methods := []*Method{}
	db.View(func(txn *badger.Txn) error {
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

func getClassConsts(db *badger.DB, scope string, name string) []*ClassConst {
	prefix := []byte("classConst" + collectionSep + scope + KeySep + name + KeySep)
	classConsts := []*ClassConst{}
	db.View(func(txn *badger.Txn) error {
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

func getProperties(db *badger.DB, scope string, name string) []*Property {
	prefix := []byte("property" + collectionSep + scope + KeySep + name + KeySep)
	properties := []*Property{}
	db.View(func(txn *badger.Txn) error {
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
