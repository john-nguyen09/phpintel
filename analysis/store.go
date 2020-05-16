package analysis

import (
	"bytes"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/stub"
	"github.com/john-nguyen09/phpintel/util"
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

	completionDataCollection  string = "completionDataCollection"
	functionCompletionIndex   string = "functionCompletionIndex"
	constCompletionIndex      string = "constCompletionIndex"
	defineCompletionIndex     string = "defineCompletionIndex"
	classCompletionIndex      string = "classCompletionIndex"
	interfaceCompletionIndex  string = "interfaceCompletionindex"
	traitCompletionIndex      string = "traitCompletionIndex"
	methodCompletionIndex     string = "methodCompletionIndex"
	propertyCompletionIndex   string = "propertyCompletionIndex"
	classConstCompletionIndex string = "classConstCompletionIndex"
	namespaceCompletionIndex  string = "namespaceCompletionIndex"
)

var /* const */ versionKey = []byte("Version")

const scopeSep = "::"

// KeySep is the separator when constructing key
const KeySep string = "\x00"

type entry struct {
	key []byte
	e   *storage.Encoder
}

func newEntry(collection string, key string) *entry {
	return &entry{
		key: []byte(collection + KeySep + key),
		e:   storage.NewEncoder(),
	}
}

func (s *entry) getEncoder() *storage.Encoder {
	return s.e
}

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) bytes() []byte {
	return s.e.Bytes()
}

type Store struct {
	uri       protocol.DocumentURI
	db        storage.DB
	fEngine   *fuzzyEngine
	stubbers  []stub.Stubber
	documents cmap.ConcurrentMap

	syncedDocumentURIs cmap.ConcurrentMap
}

type symbolDeletor struct {
	uri     string
	symbols map[string]bool
}

func newSymbolDeletor(db storage.DB, uri string) *symbolDeletor {
	entry := newEntry(documentSymbols, uri+KeySep)
	deletor := &symbolDeletor{
		uri:     uri,
		symbols: map[string]bool{},
	}
	db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		deletor.symbols[strings.Join(keyInfo[2:], KeySep)] = true
	})
	return deletor
}

func (d *symbolDeletor) MarkNotDelete(ser serialisable) {
	delete(d.symbols, ser.GetCollection()+KeySep+ser.GetKey())
}

func (d *symbolDeletor) Delete(batch storage.Batch) {
	for key := range d.symbols {
		batch.Delete([]byte(key))
		batch.Delete([]byte(documentSymbols + KeySep + d.uri + key))
	}
}

type SearchOptions struct {
	predicates []func(s Symbol) bool
	limiter    func() bool
}

func NewSearchOptions() SearchOptions {
	return SearchOptions{}
}

func (s SearchOptions) WithPredicate(predicate func(s Symbol) bool) SearchOptions {
	s.predicates = append(s.predicates, predicate)
	return s
}

func (s SearchOptions) WithLimit(limit int) SearchOptions {
	count := 0
	s.limiter = func() bool {
		count++
		return count >= limit
	}
	return s
}

func (s SearchOptions) IsLimitReached() bool {
	if s.limiter == nil {
		return false
	}
	return s.limiter()
}

func NewStore(uri protocol.DocumentURI, storePath string) (*Store, error) {
	db, err := storage.NewDisk(storePath)
	stubbers := stub.GetStubbers()
	if err != nil {
		return nil, err
	}
	store := &Store{
		uri:       uri,
		db:        db,
		fEngine:   newFuzzyEngine(db),
		stubbers:  stubbers,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}
	return store, nil
}

func (s *Store) Close() {
	s.fEngine.close()
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
	for _, stubber := range s.stubbers {
		stubber.Walk(func(path string, data []byte) error {
			document := NewDocument(stubber.GetURI(path), data)
			currentMD5 := document.GetHash()
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
		filePath, err := putil.UriToPath(uri)
		if err != nil {
			log.Printf("GetOrCreateDocument error: %v", err)
			return nil
		}
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("GetOrCreateDocument error: %v", err)
			return nil
		}
		document = NewDocument(uri, data)
		s.SaveDocOnStore(document)
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
	err := s.db.WriteBatch(func(b storage.Batch) error {
		ciDeletor := newFuzzyEngineDeletor(s.fEngine, uri)
		ciDeletor.delete()
		syDeletor := newSymbolDeletor(s.db, uri)
		syDeletor.Delete(b)
		entry := newEntry(documentCollection, uri)
		b.Delete(entry.getKeyBytes())
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

func (s *Store) DeleteFolder(uri protocol.DocumentURI) {
	entry := newEntry(documentCollection, uri)
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
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

	currentMD5 := document.GetHash()
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
	defer util.TimeTrack(time.Now(), "SyncDocument")
	err := s.db.WriteBatch(func(b storage.Batch) error {
		ciDeletor := newFuzzyEngineDeletor(s.fEngine, document.GetURI())
		syDeletor := newSymbolDeletor(s.db, document.GetURI())

		s.writeAllSymbols(b, document, ciDeletor, syDeletor)

		ciDeletor.delete()
		syDeletor.Delete(b)
		entry := newEntry(documentCollection, document.GetURI())
		b.Put(entry.getKeyBytes(), document.GetHash())
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

func (s *Store) SaveDocOnStore(document *Document) {
	s.documents.Set(document.GetURI(), document)
}

func (s *Store) PrepareForIndexing() {
	defer util.TimeTrack(time.Now(), "PrepareForIndexing")
	syncedDocumentURIs := s.getSyncedDocumentURIs()
	for key, value := range syncedDocumentURIs {
		s.syncedDocumentURIs.Set(key, value)
	}
}

func (s *Store) FinishIndexing() {
	err := s.db.WriteBatch(func(wb storage.Batch) error {
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
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		documentURIs[strings.Split(string(it.Key()), KeySep)[1]] = it.Value()
	})
	return documentURIs
}

func (s *Store) writeAllSymbols(batch storage.Batch, document *Document,
	ciDeletor *fuzzyEngineDeletor, syDeletor *symbolDeletor) {
	for _, impTable := range document.importTables {
		is := indexablesFromNamespaceName(impTable.GetNamespace())
		for index, i := range is {
			key := i.key + KeySep + strconv.Itoa(index)
			s.indexName(batch, document, i, key)
		}
	}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, child Symbol) {
		if ser, ok := child.(serialisable); ok {
			key := ser.GetKey()
			if key == "" {
				return
			}
			entry := newEntry(ser.GetCollection(), key)
			ser.Serialise(entry.e)
			writeEntry(batch, entry)
			rememberSymbol(batch, document, ser)
			syDeletor.MarkNotDelete(ser)

			if indexable, ok := child.(NameIndexable); ok {
				s.indexName(batch, document, indexable, key)
			}
		}
	})
}

func rememberSymbol(batch storage.Batch, document *Document, ser serialisable) {
	entry := newEntry(documentSymbols, document.GetURI()+KeySep+ser.GetCollection()+KeySep+ser.GetKey())
	writeEntry(batch, entry)
}

func (s *Store) indexName(batch storage.Batch, document *Document, indexable NameIndexable, key string) {
	s.fEngine.index(document.GetURI(), indexable, key)
}

func writeEntry(batch storage.Batch, entry *entry) {
	batch.Put(entry.getKeyBytes(), entry.bytes())
}

func deleteEntry(batch storage.Batch, entry *entry) {
	batch.Delete(entry.getKeyBytes())
}

func (s *Store) GetURI() protocol.DocumentURI {
	return s.uri
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

func namespacePredicate(scope string) func(s Symbol) bool {
	if scope == "" {
		return func(s Symbol) bool {
			return true
		}
	}
	return func(s Symbol) bool {
		symbolScope := ""
		if v, ok := s.(HasScope); ok {
			symbolScope = v.GetScope()
		}
		return symbolScope == scope
	}
}

func (s *Store) SearchNamespaces(keyword string, options SearchOptions) ([]string, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	// In namespace normally there isn't \ but somehow it has ignores in because
	// namespaces are not indexed with \
	if scope == "\\" {
		scope = ""
	}
	namespaces := []string{}
	query := searchQuery{
		collection: namespaceCompletionIndex + KeySep + scope,
		keyword:    keyword,
		onData: func(cv CompletionValue) onDataResult {
			parts := strings.Split(string(cv), KeySep)
			namespaces = append(namespaces, parts[0])
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return namespaces, result
}

func (s *Store) GetClasses(name string) []*Class {
	entry := newEntry(classCollection, name+KeySep)
	classes := []*Class{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classes = append(classes, ReadClass(d))
	})
	return classes
}

func (s *Store) GetClassesByScopeStream(scope string, onData func(*Class) onDataResult) {
	if scope[len(scope)-1] != '\\' {
		scope += "\\"
	}
	entry := newEntry(classCollection, scope)
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		class := ReadClass(storage.NewDecoder(it.Value()))
		result := onData(class)
		if result.shouldStop {
			it.Stop()
		}
	})
}

func (s *Store) SearchClasses(keyword string, options SearchOptions) ([]*Class, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	classes := []*Class{}
	if scope != "" {
		if keyword != "" {
			options = options.WithPredicate(func(s Symbol) bool {
				if v, ok := s.(*Class); ok {
					return strings.Contains(v.Name.GetOriginal(), keyword)
				}
				return false
			})
		}
		s.GetClassesByScopeStream(scope, func(class *Class) onDataResult {
			if isSymbolValid(class, options) {
				classes = append(classes, class)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		})
		return classes, SearchResult{true}
	}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: classCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(classCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			class := ReadClass(d)
			if isSymbolValid(class, options) {
				classes = append(classes, class)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return classes, result
}

func (s *Store) GetInterfaces(name string) []*Interface {
	entry := newEntry(interfaceCollection, name+KeySep)
	interfaces := []*Interface{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		interfaces = append(interfaces, ReadInterface(d))
	})
	return interfaces
}

func (s *Store) SearchInterfaces(keyword string, options SearchOptions) ([]*Interface, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	interfaces := []*Interface{}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: interfaceCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(interfaceCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			theInterface := ReadInterface(d)
			if isSymbolValid(theInterface, options) {
				interfaces = append(interfaces, theInterface)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return interfaces, result
}

func (s *Store) GetTraits(name string) []*Trait {
	entry := newEntry(traitCollection, name+KeySep)
	traits := []*Trait{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		traits = append(traits, ReadTrait(d))
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
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: traitCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(traitCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			trait := ReadTrait(d)
			if isSymbolValid(trait, options) {
				traits = append(traits, trait)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return traits, result
}

func (s *Store) GetFunctions(name string) []*Function {
	entry := newEntry(functionCollection, name+KeySep)
	functions := []*Function{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		functions = append(functions, ReadFunction(d))
	})
	return functions
}

func (s *Store) SearchFunctions(keyword string, options SearchOptions) ([]*Function, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	functions := []*Function{}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: functionCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(functionCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			function := ReadFunction(d)
			if isSymbolValid(function, options) {
				functions = append(functions, function)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return functions, result
}

func (s *Store) GetConsts(name string) []*Const {
	entry := newEntry(constCollection, name+KeySep)
	consts := []*Const{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		consts = append(consts, ReadConst(d))
	})
	return consts
}

func (s *Store) SearchConsts(keyword string, options SearchOptions) ([]*Const, SearchResult) {
	consts := []*Const{}
	query := searchQuery{
		collection: constCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(constCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			constant := ReadConst(d)
			if isSymbolValid(constant, options) {
				consts = append(consts, constant)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return consts, result
}

func (s *Store) GetDefines(name string) []*Define {
	entry := newEntry(defineCollection, name+KeySep)
	defines := []*Define{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		defines = append(defines, ReadDefine(d))
	})
	return defines
}

func (s *Store) SearchDefines(keyword string, options SearchOptions) ([]*Define, SearchResult) {
	defines := []*Define{}
	query := searchQuery{
		collection: defineCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(defineCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			define := ReadDefine(d)
			if isSymbolValid(define, options) {
				defines = append(defines, define)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return defines, result
}

func (s *Store) GetMethods(scope string, name string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep+name+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		methods = append(methods, ReadMethod(d))
	})
	return methods
}

func (s *Store) GetAllMethods(scope string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		methods = append(methods, ReadMethod(d))
	})
	return methods
}

func (s *Store) SearchMethods(scope string, keyword string, options SearchOptions) ([]*Method, SearchResult) {
	if keyword == "" {
		return []*Method{}, SearchResult{false}
	}

	methods := []*Method{}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: methodCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(methodCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			method := ReadMethod(d)
			if isSymbolValid(method, options) {
				methods = append(methods, method)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return methods, result
}

func (s *Store) GetClassConsts(scope string, name string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep+name)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classConsts = append(classConsts, ReadClassConst(d))
	})
	return classConsts
}

func (s *Store) GetAllClassConsts(scope string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classConsts = append(classConsts, ReadClassConst(d))
	})
	return classConsts
}

func (s *Store) SearchClassConsts(scope string, keyword string, options SearchOptions) ([]*ClassConst, SearchResult) {
	if keyword == "" {
		return s.GetAllClassConsts(scope), SearchResult{true}
	}

	classConsts := []*ClassConst{}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: classConstCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(classConstCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			classConst := ReadClassConst(d)
			if isSymbolValid(classConst, options) {
				classConsts = append(classConsts, classConst)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return classConsts, result
}

func (s *Store) GetProperties(scope string, name string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep+name+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		properties = append(properties, ReadProperty(d))
	})
	return properties
}

func (s *Store) GetAllProperties(scope string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		properties = append(properties, ReadProperty(d))
	})
	return properties
}

func (s *Store) SearchProperties(scope string, keyword string, options SearchOptions) ([]*Property, SearchResult) {
	if keyword == "" {
		return s.GetAllProperties(scope), SearchResult{true}
	}

	properties := []*Property{}
	options.predicates = append(options.predicates, namespacePredicate(scope))
	query := searchQuery{
		collection: propertyCompletionIndex,
		keyword:    keyword,
		onData: func(completionValue CompletionValue) onDataResult {
			entry := newEntry(propertyCollection, string(completionValue))
			value, err := s.db.Get(entry.getKeyBytes())
			if err != nil {
				return onDataResult{false}
			}
			d := storage.NewDecoder(value)
			property := ReadProperty(d)
			if isSymbolValid(property, options) {
				properties = append(properties, property)
				if options.IsLimitReached() {
					return onDataResult{true}
				}
			}
			return onDataResult{false}
		},
	}
	result := s.fEngine.search(query)
	return properties, result
}

func (s *Store) GetGlobalVariables(name string) []*GlobalVariable {
	entry := newEntry(globalVariableCollection, name+KeySep)
	results := []*GlobalVariable{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		results = append(results, ReadGlobalVariable(d))
	})
	return results
}
