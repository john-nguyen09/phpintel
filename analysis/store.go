package analysis

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/jmhodges/levigo"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	putil "github.com/john-nguyen09/phpintel/util"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	documentSymbols          string = "documentSymbols"
	classCollection          string = "class"
	interfaceCollection      string = "interface"
	traitCollection          string = "trait"
	functionCollection       string = "function"
	constCollection          string = "const"
	defineCollection         string = "define"
	methodCollection         string = "method"
	classConstCollection     string = "classConst"
	propertyCollection       string = "property"
	globalVariableCollection string = "globalVariable"
	documentCollection       string = "document"

	documentCompletionIndex   string = "documentCompletionIndices"
	functionCompletionIndex   string = "functionCompletionIndex"
	constCompletionIndex      string = "constCompletionIndex"
	defineCompletionIndex     string = "defineCompletionIndex"
	classCompletionIndex      string = "classCompletionIndex"
	interfaceCompletionIndex  string = "interfaceCompletionindex"
	traitCompletionIndex      string = "traitCompletionIndex"
	methodCompletionIndex     string = "methodCompletionIndex"
	propertyCompletionIndex   string = "propertyCompletionIndex"
	classConstCompletionIndex string = "classConstCompletionIndex"
)

var /* const */ versionKey = []byte("Version")

const scopeSep = "::"

var stubs = []*PhpStub{}

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

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) getBytes() []byte {
	return s.serialiser.GetBytes()
}

type Store struct {
	uri       protocol.DocumentURI
	db        *storage.Storage
	documents cmap.ConcurrentMap

	syncedDocumentURIs cmap.ConcurrentMap
}

type symbolDeletor struct {
	uri     string
	symbols map[string]bool
}

func newSymbolDeletor(db *storage.Storage, uri string) *symbolDeletor {
	entry := newEntry(documentSymbols, uri+KeySep)
	deletor := &symbolDeletor{
		uri:     uri,
		symbols: map[string]bool{},
	}
	db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		deletor.symbols[strings.Join(keyInfo[2:], KeySep)] = true
	})
	return deletor
}

func (d *symbolDeletor) MarkNotDelete(ser serialisable) {
	delete(d.symbols, ser.GetCollection()+KeySep+ser.GetKey())
}

func (d *symbolDeletor) Delete(batch *levigo.WriteBatch) {
	for key := range d.symbols {
		batch.Delete([]byte(key))
		batch.Delete([]byte(documentSymbols + KeySep + d.uri + key))
	}
}

type SearchOptions struct {
	predicates []func(s Symbol) bool
	limit      int
}

func NewSearchOptions() SearchOptions {
	return SearchOptions{}
}

func (s SearchOptions) WithPredicate(predicate func(s Symbol) bool) SearchOptions {
	s.predicates = append(s.predicates, predicate)
	return s
}

func (s SearchOptions) WithLimit(limit int) SearchOptions {
	s.limit = limit
	return s
}

func initStubs() {
	stubs = append(stubs, NewPhpStormStub())
}

func NewStore(uri protocol.DocumentURI, storePath string) (*Store, error) {
	db, err := storage.NewStorage(storePath)
	initStubs()
	if err != nil {
		return nil, err
	}
	store := &Store{
		uri:       uri,
		db:        db,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}
	return store, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) GetStoreVersion() string {
	v, err := s.db.Get(versionKey)
	if err != nil {
		return "v0.0.0"
	}
	return string(v)
}

func (s *Store) PutVersion(version string) {
	s.db.Put(versionKey, []byte(version))
}

func (s *Store) Clear() {
	s.db.Clear()
}

func (s *Store) Migrate(newVersion string) {
	storeVersion := s.GetStoreVersion()
	sv, _ := semver.NewVersion(storeVersion)

	if sv == nil {
		return
	}

	targetV, _ := semver.NewVersion("v0.0.12")
	if sv.LessThan(targetV) {
		log.Println("Clearing database for upgrade.")
		s.Clear()
		s.PutVersion(newVersion)
	}
}

func (s *Store) LoadStubs() {
	for _, stub := range stubs {
		stub.Walk(func(path string, data []byte) error {
			document := NewDocument(stub.GetUri(path), string(data))
			currentMD5 := document.GetMD5Hash()
			entry := newEntry(documentCollection, document.GetURI())
			savedMD5, err := s.db.Get(entry.getKeyBytes())
			if err != nil || bytes.Compare(currentMD5, savedMD5) != 0 {
				document.Load()
				s.SyncDocument(document)
			}
			return nil
		})
	}
}

func (s *Store) GetOrCreateDocument(uri protocol.DocumentURI) *Document {
	var document *Document
	if value, ok := s.documents.Get(uri); !ok {
		filePath := putil.UriToPath(uri)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil
		}
		document = NewDocument(uri, string(data))
		s.documents.Set(uri, document)
	} else {
		document = value.(*Document)
	}
	return document
}

func (s *Store) OpenDocument(uri protocol.DocumentURI) *Document {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("Document %s not found", uri)
		return nil
	}
	document.Lock()
	defer func() {
		document.Unlock()
		s.releaseDocIfNotOpen(document)
	}()
	document.Open()
	document.Load()
	s.SyncDocument(document)
	return document
}

func (s *Store) CloseDocument(uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("document %s not found", uri)
		return
	}
	document.Lock()
	defer func() {
		document.Unlock()
		s.releaseDocIfNotOpen(document)
	}()
	document.Close()
	s.SyncDocument(document)
}

func (s *Store) DeleteDocument(uri protocol.DocumentURI) {
	err := s.db.WriteBatch(func(wb *levigo.WriteBatch) error {
		ciDeletor := newCompletionIndexDeletor(s.db, uri)
		ciDeletor.Delete(wb)
		syDeletor := newSymbolDeletor(s.db, uri)
		syDeletor.Delete(wb)
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (s *Store) DeleteFolder(uri protocol.DocumentURI) {
	entry := newEntry(documentCollection, uri)
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		uri := strings.Split(string(it.Key()), KeySep)[1]
		s.DeleteDocument(uri)
	})
}

func (s *Store) CompareAndIndexDocument(filePath string) *Document {
	uri := putil.PathToUri(filePath)
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		return nil
	}
	document.Lock()
	defer func() {
		document.Unlock()
		s.releaseDocIfNotOpen(document)
	}()

	currentMD5 := document.GetMD5Hash()
	savedMD5, ok := s.syncedDocumentURIs.Get(uri)
	if ok {
		s.syncedDocumentURIs.Remove(uri)
	} else {
		entry := newEntry(documentCollection, document.GetURI())
		value, err := s.db.Get(entry.getKeyBytes())
		if err == nil {
			savedMD5 = value
		}
	}
	if savedMD5 != nil && bytes.Compare(currentMD5, savedMD5.([]byte)) == 0 {
		return document
	}

	document.Load()
	s.SyncDocument(document)
	return document
}

func (s *Store) SyncDocument(document *Document) {
	err := s.db.WriteBatch(func(wb *levigo.WriteBatch) error {
		ciDeletor := newCompletionIndexDeletor(s.db, document.GetURI())
		syDeletor := newSymbolDeletor(s.db, document.GetURI())

		s.writeAllSymbols(wb, document, ciDeletor, syDeletor)

		ciDeletor.Delete(wb)
		syDeletor.Delete(wb)
		entry := newEntry(documentCollection, document.GetURI())
		wb.Put(entry.getKeyBytes(), document.GetMD5Hash())
		return nil
	})
	if err != nil {
		log.Print(err)
	}
}

func (s *Store) releaseDocIfNotOpen(document *Document) {
	if !document.IsOpen() {
		s.documents.Remove(document.uri)
	}
}

func (s *Store) PrepareForIndexing() {
	syncedDocumentURIs := s.getSyncedDocumentURIs()
	for key, value := range syncedDocumentURIs {
		s.syncedDocumentURIs.Set(key, value)
	}
}

func (s *Store) FinishIndexing() {
	err := s.db.WriteBatch(func(wb *levigo.WriteBatch) error {
		for iter := range s.syncedDocumentURIs.Iter() {
			s.DeleteDocument(iter.Key)
			s.syncedDocumentURIs.Remove(iter.Key)
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (s *Store) getSyncedDocumentURIs() map[string][]byte {
	documentURIs := make(map[string][]byte)
	entry := newEntry(documentCollection, "file://")
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		documentURIs[strings.Split(string(it.Key()), KeySep)[1]] = it.Value()
	})
	return documentURIs
}

func (s *Store) forgetAllSymbols(batch *levigo.WriteBatch, uri string) {
	entry := newEntry(documentSymbols, uri+KeySep)
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		toBeDelete := newEntry(keyInfo[2], strings.Join(keyInfo[3:], KeySep))
		deleteEntry(batch, toBeDelete)
	})
}

func (s *Store) writeAllSymbols(batch *levigo.WriteBatch, document *Document,
	ciDeletor *completionIndexDeletor, syDeletor *symbolDeletor) {
	for _, child := range document.Children {
		if ser, ok := child.(serialisable); ok {
			key := ser.GetKey()
			if key == "" {
				continue
			}
			entry := newEntry(ser.GetCollection(), key)
			ser.Serialise(entry.serialiser)
			writeEntry(batch, entry)
			rememberSymbol(batch, document, ser)
			syDeletor.MarkNotDelete(ser)

			if indexable, ok := child.(NameIndexable); ok {
				indexName(batch, document, indexable, key)
				ciDeletor.MarkNotDelete(document.GetURI(), indexable, key)
			}
		}
	}
}

func rememberSymbol(batch *levigo.WriteBatch, document *Document, ser serialisable) {
	entry := newEntry(documentSymbols, document.GetURI()+KeySep+ser.GetCollection()+KeySep+ser.GetKey())
	writeEntry(batch, entry)
}

func indexName(batch *levigo.WriteBatch, document *Document, indexable NameIndexable, key string) {
	entries := createCompletionEntries(document.GetURI(), indexable, key)
	for _, entry := range entries {
		writeEntry(batch, entry)
	}
}

func writeEntry(batch *levigo.WriteBatch, entry *entry) {
	batch.Put(entry.getKeyBytes(), entry.getBytes())
}

func deleteEntry(batch *levigo.WriteBatch, entry *entry) {
	batch.Delete(entry.getKeyBytes())
}

func (s *Store) GetURI() protocol.DocumentURI {
	return s.uri
}

func (s *Store) GetClasses(name string) []*Class {
	entry := newEntry(classCollection, name+KeySep)
	classes := []*Class{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		classes = append(classes, ReadClass(serialiser))
	})
	return classes
}

func isSymbolValid(symbol Symbol, options SearchOptions) bool {
	if len(options.predicates) == 0 {
		return true
	}
	allTrue := true
	for _, predicate := range options.predicates {
		if !predicate(symbol) {
			allTrue = false
			break
		}
	}
	return allTrue
}

func (s *Store) SearchClasses(keyword string, options SearchOptions) ([]*Class, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	prefixes := []string{""}
	if scope != "" {
		prefixes = append(prefixes, scope)
	}
	classes := []*Class{}
	count := 0
	query := searchQuery{
		collection: classCompletionIndex,
		prefixes:   prefixes,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(classCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			class := ReadClass(serialiser)
			if isSymbolValid(class, options) {
				classes = append(classes, class)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return classes, result
}

func (s *Store) GetInterfaces(name string) []*Interface {
	entry := newEntry(interfaceCollection, name+KeySep)
	interfaces := []*Interface{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		interfaces = append(interfaces, ReadInterface(serialiser))
	})
	return interfaces
}

func (s *Store) SearchInterfaces(keyword string, options SearchOptions) ([]*Interface, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	prefixes := []string{""}
	if scope != "" {
		prefixes = append(prefixes, scope)
	}
	interfaces := []*Interface{}
	count := 0
	query := searchQuery{
		collection: interfaceCompletionIndex,
		prefixes:   prefixes,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(interfaceCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			theInterface := ReadInterface(serialiser)
			if isSymbolValid(theInterface, options) {
				interfaces = append(interfaces, theInterface)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return interfaces, result
}

func (s *Store) GetTraits(name string) []*Trait {
	entry := newEntry(traitCollection, name+KeySep)
	traits := []*Trait{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		traits = append(traits, ReadTrait(serialiser))
	})
	return traits
}

func (s *Store) SearchTraits(keyword string, options SearchOptions) ([]*Trait, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	prefixes := []string{""}
	if scope != "" {
		prefixes = append(prefixes, scope)
	}
	traits := []*Trait{}
	count := 0
	query := searchQuery{
		collection: traitCompletionIndex,
		prefixes:   prefixes,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(traitCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			trait := ReadTrait(serialiser)
			if isSymbolValid(trait, options) {
				traits = append(traits, trait)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return traits, result
}

func (s *Store) GetFunctions(name string) []*Function {
	entry := newEntry(functionCollection, name+KeySep)
	functions := []*Function{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		functions = append(functions, ReadFunction(serialiser))
	})
	return functions
}

func (s *Store) SearchFunctions(keyword string, options SearchOptions) ([]*Function, SearchResult) {
	functions := []*Function{}
	count := 0
	query := searchQuery{
		collection: functionCompletionIndex,
		prefixes:   []string{""},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(functionCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			function := ReadFunction(serialiser)
			if isSymbolValid(function, options) {
				functions = append(functions, function)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return functions, result
}

func (s *Store) GetConsts(name string) []*Const {
	entry := newEntry(constCollection, name+KeySep)
	consts := []*Const{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		consts = append(consts, ReadConst(serialiser))
	})
	return consts
}

func (s *Store) SearchConsts(keyword string, options SearchOptions) ([]*Const, SearchResult) {
	consts := []*Const{}
	count := 0
	query := searchQuery{
		collection: constCompletionIndex,
		prefixes:   []string{},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(constCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			constant := ReadConst(serialiser)
			if isSymbolValid(constant, options) {
				consts = append(consts, constant)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return consts, result
}

func (s *Store) GetDefines(name string) []*Define {
	entry := newEntry(defineCollection, name+KeySep)
	defines := []*Define{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		defines = append(defines, ReadDefine(serialiser))
	})
	return defines
}

func (s *Store) SearchDefines(keyword string, options SearchOptions) ([]*Define, SearchResult) {
	defines := []*Define{}
	count := 0
	query := searchQuery{
		collection: defineCompletionIndex,
		prefixes:   []string{""},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(defineCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			define := ReadDefine(serialiser)
			if isSymbolValid(define, options) {
				defines = append(defines, define)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return defines, result
}

func (s *Store) GetMethods(scope string, name string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep+name+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		methods = append(methods, ReadMethod(serialiser))
	})
	return methods
}

func (s *Store) GetAllMethods(scope string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		methods = append(methods, ReadMethod(serialiser))
	})
	return methods
}

func (s *Store) SearchMethods(scope string, keyword string, options SearchOptions) ([]*Method, SearchResult) {
	if keyword == "" {
		return []*Method{}, SearchResult{false}
	}

	methods := []*Method{}
	count := 0
	query := searchQuery{
		collection: methodCompletionIndex,
		prefixes:   []string{scope},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(methodCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			method := ReadMethod(serialiser)
			if isSymbolValid(method, options) {
				methods = append(methods, method)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return methods, result
}

func (s *Store) GetClassConsts(scope string, name string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep+name)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		classConsts = append(classConsts, ReadClassConst(serialiser))
	})
	return classConsts
}

func (s *Store) GetAllClassConsts(scope string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		classConsts = append(classConsts, ReadClassConst(serialiser))
	})
	return classConsts
}

func (s *Store) SearchClassConsts(scope string, keyword string, options SearchOptions) ([]*ClassConst, SearchResult) {
	if keyword == "" {
		return s.GetAllClassConsts(scope), SearchResult{true}
	}

	classConsts := []*ClassConst{}
	count := 0
	query := searchQuery{
		collection: classConstCompletionIndex,
		prefixes:   []string{scope},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(classConstCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			classConst := ReadClassConst(serialiser)
			if isSymbolValid(classConst, options) {
				classConsts = append(classConsts, classConst)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return classConsts, result
}

func (s *Store) GetProperties(scope string, name string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep+name+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		properties = append(properties, ReadProperty(serialiser))
	})
	return properties
}

func (s *Store) GetAllProperties(scope string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		properties = append(properties, ReadProperty(serialiser))
	})
	return properties
}

func (s *Store) SearchProperties(scope string, keyword string, options SearchOptions) ([]*Property, SearchResult) {
	if keyword == "" {
		return s.GetAllProperties(scope), SearchResult{true}
	}

	properties := []*Property{}
	count := 0
	query := searchQuery{
		collection: propertyCompletionIndex,
		prefixes:   []string{scope},
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(propertyCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			serialiser := SerialiserFromByteSlice(value)
			property := ReadProperty(serialiser)
			if isSymbolValid(property, options) {
				properties = append(properties, property)
				count++
			}
			if options.limit > 0 && count >= options.limit {
				return onDataResult{true}
			}
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.db, query)
	return properties, result
}

func (s *Store) GetGlobalVariables(name string) []*GlobalVariable {
	entry := newEntry(globalVariableCollection, name+KeySep)
	results := []*GlobalVariable{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		serialiser := SerialiserFromByteSlice(it.Value())
		results = append(results, ReadGlobalVariable(serialiser))
	})
	return results
}
