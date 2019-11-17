package analysis

import (
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	putil "github.com/john-nguyen09/phpintel/util"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	documentSymbols      string = "documentSymbols"
	classCollection      string = "class"
	interfaceCollection  string = "interface"
	traitCollection      string = "trait"
	functionCollection   string = "function"
	constCollection      string = "const"
	defineCollection     string = "define"
	methodCollection     string = "method"
	classConstCollection string = "classConst"
	propertyCollection   string = "property"
)

// KeySep is the separator when constructing key
const KeySep string = "\x00"

type entry struct {
	key        []byte
	serialiser *Serialiser
}

func newEntry(collection string, key string) *entry {
	return &entry{
		key:        []byte(collection + KeySep + key),
		serialiser: NewSerialiser(),
	}
}

func (s *entry) getSerialiser() *Serialiser {
	return s.serialiser
}

func (s *entry) getKeyRange() *util.Range {
	return &util.Range{
		Start: append(s.getKeyBytes(), '\x00'),
		Limit: append(s.getKeyBytes(), '\xFF'),
	}
}

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) getBytes() []byte {
	return s.serialiser.GetBytes()
}

type Store struct {
	db            *leveldb.DB
	documentLocks map[string]sync.Mutex
	documentMu    sync.Mutex
	documents     map[string]*Document
}

func NewStore(storePath string) (*Store, error) {
	db, err := leveldb.OpenFile(storePath, nil)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:            db,
		documentLocks: map[string]sync.Mutex{},
		documents:     map[string]*Document{},
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetOrCreateDocument(uri protocol.DocumentURI) *Document {
	var document *Document
	var ok bool
	s.documentMu.Lock()
	defer s.documentMu.Unlock()
	if document, ok = s.documents[uri]; !ok {
		filePath := putil.UriToPath(uri)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("cannot read %s, error: %s", filePath, err)
			return nil
		}
		document = NewDocument(uri, string(data))
		s.documents[uri] = document
	}
	return document
}

func (s *Store) OpenDocument(uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("Document %s not found", uri)
		return
	}
	document.Open()
}

func (s *Store) CloseDocument(uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("document %s not found", uri)
		return
	}
	document.Close()
}

func (s *Store) IndexDocument(filePath string) {
	uri := putil.PathToUri(filePath)
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("Failed to get %s", filePath)
		return
	}
	document.Load()
	s.SyncDocument(document)
	if !document.isOpen {
		document.Release()
	}
}

func (s *Store) ChangeDocument(uri string, changes []protocol.TextDocumentContentChangeEvent) error {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("Document %s not found", uri)
	}
	document.ApplyChanges(changes)
	document.Release()
	document.Load()
	s.SyncDocument(document)
	if !document.isOpen {
		document.Release()
	}
	return nil
}

func (s *Store) SyncDocument(document *Document) {
	batch := new(leveldb.Batch)
	s.forgetAllSymbols(batch, document)
	s.writeAllSymbols(batch, document)
	err := s.db.Write(batch, nil)
	if err != nil {
		log.Print(err)
	}
}

func (s *Store) forgetAllSymbols(batch *leveldb.Batch, document *Document) {
	entry := newEntry(documentSymbols, document.GetURI()+KeySep)
	it := s.db.NewIterator(entry.getKeyRange(), nil)
	defer it.Release()
	for it.Next() {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		toBeDelete := newEntry(keyInfo[1], keyInfo[2])
		deleteEntry(batch, toBeDelete)
	}
}

func (s *Store) writeAllSymbols(batch *leveldb.Batch, document *Document) {
	for _, child := range document.Children {
		if serialisable, ok := child.(Serialisable); ok {
			entry := newEntry(serialisable.GetCollection(), serialisable.GetKey())
			serialisable.Serialise(entry.serialiser)
			writeEntry(batch, entry)
			rememberSymbol(batch, document, serialisable)
		}
	}
}

func rememberSymbol(batch *leveldb.Batch, document *Document, serialisable Serialisable) {
	entry := newEntry(documentSymbols, document.GetURI()+KeySep+serialisable.GetCollection()+KeySep+serialisable.GetKey())
	writeEntry(batch, entry)
}

func writeEntry(batch *leveldb.Batch, entry *entry) {
	batch.Put(entry.getKeyBytes(), entry.getBytes())
}

func deleteEntry(batch *leveldb.Batch, entry *entry) {
	batch.Delete(entry.getKeyBytes())
}

func (s *Store) GetClasses(name string) []*Class {
	prefix := []byte("class" + KeySep + name + KeySep)
	classes := []*Class{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		classes = append(classes, ReadClass(serialiser))
	}
	return classes
}

func (s *Store) GetInterfaces(name string) []*Interface {
	prefix := []byte("interface" + KeySep + name)
	interfaces := []*Interface{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		interfaces = append(interfaces, ReadInterface(serialiser))
	}
	return interfaces
}

func (s *Store) GetTraits(name string) []*Trait {
	prefix := []byte("trait" + KeySep + name)
	traits := []*Trait{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		traits = append(traits, ReadTrait(serialiser))
	}
	return traits
}

func (s *Store) GetFunctions(name string) []*Function {
	prefix := []byte("function" + KeySep + name)
	functions := []*Function{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		functions = append(functions, ReadFunction(serialiser))
	}
	return functions
}

func (s *Store) GetConsts(name string) []*Const {
	prefix := []byte("const" + KeySep + name)
	consts := []*Const{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		consts = append(consts, ReadConst(serialiser))
	}
	return consts
}

func (s *Store) GetDefines(name string) []*Define {
	prefix := []byte("define" + KeySep + name)
	defines := []*Define{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		defines = append(defines, ReadDefine(serialiser))
	}
	return defines
}

func (s *Store) GetMethods(scope string, name string) []*Method {
	prefix := []byte("method" + KeySep + scope + KeySep + name)
	methods := []*Method{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		methods = append(methods, ReadMethod(serialiser))
	}
	return methods
}

func (s *Store) GetClassConsts(scope string, name string) []*ClassConst {
	prefix := []byte("classConst" + KeySep + scope + KeySep + name)
	classConsts := []*ClassConst{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		classConsts = append(classConsts, ReadClassConst(serialiser))
	}
	return classConsts
}

func (s *Store) GetProperties(scope string, name string) []*Property {
	prefix := []byte("property" + KeySep + scope + KeySep + name)
	properties := []*Property{}
	it := s.db.NewIterator(util.BytesPrefix(prefix), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		properties = append(properties, ReadProperty(serialiser))
	}
	return properties
}
