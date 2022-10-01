package analysis

import (
	"bytes"
	"context"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/akrylysov/pogreb"
	"github.com/bep/debounce"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/stub"
	"github.com/john-nguyen09/phpintel/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/zyedidia/generic/mapset"
)

const (
	documentSymbolsCollection    string = "docSym"
	classCollection              string = "cla"
	interfaceCollection          string = "int"
	traitCollection              string = "tra"
	functionCollection           string = "fun"
	constCollection              string = "con"
	defineCollection             string = "def"
	methodCollection             string = "met"
	classConstCollection         string = "claCon"
	propertyCollection           string = "pro"
	globalVariableCollection     string = "gloVar"
	documentCollection           string = "doc"
	documentNamespacesCollection string = "docNs"

	documentCompletionIndex   string = "docCom"
	completionDataCollection  string = "comDatCol"
	functionCompletionIndex   string = "funCom"
	constCompletionIndex      string = "conCom"
	defineCompletionIndex     string = "defCom"
	classCompletionIndex      string = "claCom"
	interfaceCompletionIndex  string = "intCom"
	traitCompletionIndex      string = "traCom"
	methodCompletionIndex     string = "metCom"
	propertyCompletionIndex   string = "proCom"
	classConstCompletionIndex string = "claConCom"
	namespaceCompletionIndex  string = "namCom"

	referenceIndexCollection string = "refInd"
	documentReferenceIndex   string = "docRefInd"
)

var /* const */ versionKey = []byte("Version")

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

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) bytes() []byte {
	return s.e.Bytes()
}

type documentSymbol struct {
	collection    string
	key           string
	indexableName string
}

func (s documentSymbol) serialise(e *storage.Encoder) {
	e.WriteString(s.collection)
	e.WriteString(s.key)
	e.WriteString(s.indexableName)
}

func readDocumentSymbol(d *storage.Decoder) documentSymbol {
	return documentSymbol{
		collection:    d.ReadString(),
		key:           d.ReadString(),
		indexableName: d.ReadString(),
	}
}

type documentNamespace struct {
	fullName string
}

func (s documentNamespace) serialise(e *storage.Encoder) {
	e.WriteString(s.fullName)
}

// Store contains information about a given folder and functions
// for querying symbols
type Store struct {
	uri       protocol.DocumentURI
	FS        protocol.FS
	db        storage.DB
	comDB     storage.DB
	greb      *pogreb.DB
	refIndex  *referenceIndex
	stubbers  []stub.Stubber
	documents cmap.ConcurrentMap
	isClosed  bool

	syncedDocumentURIs   cmap.ConcurrentMap
	DebouncedDeprecation func(func())
}

func readDocumentSymbols(greb *pogreb.DB, e *entry, fn func(documentSymbol) bool) {
	b, err := greb.Get(e.getKeyBytes())
	if err == nil && len(b) > 0 {
		d := storage.NewDecoder(b)
		count := d.ReadInt()
		for i := 0; i < count; i++ {
			symbol := readDocumentSymbol(d)
			if fn(symbol) {
				break
			}
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

type symbolDeletor struct {
	uri     string
	symbols mapset.Set[string]
}

func newSymbolDeletor(db storage.DB, greb *pogreb.DB, uri string) *symbolDeletor {
	deletor := &symbolDeletor{
		uri:     uri,
		symbols: mapset.New[string](),
	}
	readDocumentSymbols(greb, newEntry(documentSymbolsCollection, uri), func(symbol documentSymbol) bool {
		deletor.symbols.Put(symbol.collection + KeySep + symbol.key)
		return false
	})
	return deletor
}

func (d *symbolDeletor) MarkNotDelete(ser serialisable) {
	d.symbols.Remove(ser.GetCollection() + KeySep + ser.GetKey())
}

func (d *symbolDeletor) Delete(batch storage.Batch) {
	d.symbols.Each(func(key string) {
		batch.Delete([]byte(key))
	})
}

// SearchOptions contains predicates and limiter if applicable for searching
// symbols
type SearchOptions struct {
	predicates []func(s Symbol) bool
	limiter    func() bool
}

// NewSearchOptions creates an empty search option
func NewSearchOptions() SearchOptions {
	return SearchOptions{}
}

// WithPredicate adds a predicate into the search option
func (s SearchOptions) WithPredicate(predicate func(s Symbol) bool) SearchOptions {
	s.predicates = append(s.predicates, predicate)
	return s
}

// WithLimit adds a limiter into the search option
func (s SearchOptions) WithLimit(limit int) SearchOptions {
	count := 0
	s.limiter = func() bool {
		count++
		return count >= limit
	}
	return s
}

// IsLimitReached returns true if the limitter reaches the limit
func (s SearchOptions) IsLimitReached() bool {
	if s.limiter == nil {
		return false
	}
	return s.limiter()
}

// NewStore creates a store with the disk storage or returns an error
// if the disk storage cannot be created
func NewStore(fs protocol.FS, uri protocol.DocumentURI, storePath string) (*Store, error) {
	_, err := os.Stat(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(storePath, os.ModePerm)
		}
	}
	db, err := storage.NewGoLevelDB(storePath)
	stubbers := stub.GetStubbers()
	if err != nil {
		return nil, err
	}
	comDB, err := storage.NewGoLevelDB(path.Join(storePath, "completion"))
	if err != nil {
		return nil, err
	}
	greb, err := pogreb.Open(path.Join(storePath, "pogreb"), nil)
	if err != nil {
		return nil, err
	}
	store := &Store{
		uri:       uri,
		FS:        fs,
		db:        db,
		comDB:     comDB,
		greb:      greb,
		refIndex:  newReferenceIndex(db),
		stubbers:  stubbers,
		documents: cmap.New(),

		syncedDocumentURIs:   cmap.New(),
		DebouncedDeprecation: debounce.New(2 * time.Second),
	}
	return store, nil
}

// Close triggers close on the fuzzy engine, and closes the disk storage
func (s *Store) Close() {
	if s.isClosed {
		return
	}
	s.db.Close()
	s.greb.Close()
	s.isClosed = true
}

// GetStoreVersion returns the version of the disk storage or v0.0.0 if
// the version is missing from the disk
func (s *Store) GetStoreVersion() string {
	v, err := s.db.Get(versionKey)
	if err != nil {
		return "v0.0.0"
	}
	return string(v)
}

// PutVersion stores the given version on the disk
func (s *Store) PutVersion(version string) {
	s.db.Put(versionKey, []byte(version))
}

// Clear triggers a disk clear
func (s *Store) Clear() {
	s.db.Clear()
}

// Migrate checks for defined version if it is less than
// clears the store
func (s *Store) Migrate(newVersion string) {
	storeVersion := s.GetStoreVersion()
	sv, _ := semver.NewVersion(storeVersion)

	if sv == nil {
		return
	}

	targetV, _ := semver.NewVersion("v0.0.13")
	if sv.LessThan(targetV) {
		log.Println("Clearing database for upgrade.")
		s.Clear()
		s.PutVersion(newVersion)
	}
}

// LoadStubs loads the defined stubs, compare their hash and index them
// if needed
func (s *Store) LoadStubs() {
	start := time.Now()
	for _, stubber := range s.stubbers {
		stubber.Walk(func(path string, data []byte) error {
			document := NewDocument(stubber.GetURI(path), data)
			currentMD5 := document.GetHash()
			entry := newEntry(documentCollection, document.GetURI())
			savedMD5, err := s.db.Get(entry.getKeyBytes())
			if err != nil || !bytes.Equal(currentMD5, savedMD5) {
				document.Load()
				s.SyncDocument(document)
			}
			return nil
		})
	}
	log.Printf("LoadStubs took %s", time.Since(start))
}

// GetOrCreateDocument checks if the store contains the given URI or
// create a new document with the given URI
func (s *Store) GetOrCreateDocument(ctx context.Context, uri protocol.DocumentURI) *Document {
	var document *Document
	if value, ok := s.documents.Get(uri); !ok {
		data, err := s.FS.ReadFile(ctx, uri)
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

// OpenDocument loads and index the document with the given URI, at the same time
// marks it as open to retain it on the memory
func (s *Store) OpenDocument(ctx context.Context, uri protocol.DocumentURI) *Document {
	document := s.GetOrCreateDocument(ctx, uri)
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

// CloseDocument syncs the document with the given URI and marks
// it as close to release memory
func (s *Store) CloseDocument(ctx context.Context, uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(ctx, uri)
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

// DeleteDocument deletes all symbols and indexes relating to the URI
// this returns error if the disk cannot be written
func (s *Store) DeleteDocument(uri protocol.DocumentURI) {
	err := s.db.WriteBatch(func(b storage.Batch) error {
		ciDeletor := newCompletionIndexDeletor(s.db, uri)
		ciDeletor.Delete(b)
		syDeletor := newSymbolDeletor(s.db, s.greb, uri)
		syDeletor.Delete(b)
		s.refIndex.deleteURI(s, b, uri)
		entry := newEntry(documentCollection, uri)
		b.Delete(entry.getKeyBytes())
		return nil
	})
	if err != nil {
		log.Println(err)
	}
}

// DeleteFolder searches for documents and triggers `DeleteDocument`
func (s *Store) DeleteFolder(uri protocol.DocumentURI) {
	entry := newEntry(documentCollection, uri)
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		uri := strings.Split(string(it.Key()), KeySep)[1]
		s.DeleteDocument(uri)
	})
}

// CompareAndIndexDocument compares the file's hash with the stored one
// on the disk, and if they are not matched load the document and sync.
// The pointer to the document is returned
func (s *Store) CompareAndIndexDocument(ctx context.Context, uri string) *Document {
	document := s.GetOrCreateDocument(ctx, uri)
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
	if savedMD5 != nil && bytes.Equal(currentMD5, savedMD5.([]byte)) {
		return document
	}

	document.Load()
	s.SyncDocument(document)
	return document
}

// SyncDocument writes all definition symbols and indexes to the disk
// or the fuzzy engine
func (s *Store) SyncDocument(document *Document) {
	defer util.TimeTrack(time.Now(), "SyncDocument")
	err := s.db.WriteBatch(func(b storage.Batch) error {
		s.comDB.WriteBatch(func(comB storage.Batch) error {
			ciDeletor := newCompletionIndexDeletor(s.comDB, document.GetURI())
			syDeletor := newSymbolDeletor(s.db, s.greb, document.GetURI())
			s.refIndex.deleteURI(s, b, document.GetURI())

			s.writeAllSymbols(b, comB, document, ciDeletor, syDeletor)

			syDeletor.Delete(b)
			entry := newEntry(documentCollection, document.GetURI())
			b.Put(entry.getKeyBytes(), document.GetHash())
			return nil
		})
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

// SaveDocOnStore retains the document in memory
func (s *Store) SaveDocOnStore(document *Document) {
	s.documents.Set(document.GetURI(), document)
}

// PrepareForIndexing loads all the synced documents from the disk storage
// into memory, this is to make sure that any deleted documents that are
// not yet synced, get deleted
func (s *Store) PrepareForIndexing() {
	defer util.TimeTrack(time.Now(), "PrepareForIndexing")
	syncedDocumentURIs := s.getSyncedDocumentURIs()
	for key, value := range syncedDocumentURIs {
		s.syncedDocumentURIs.Set(key, value)
	}
}

// FinishIndexing checks for all URIs that are not removed from the map
// and delete them, because if the file exists its URI should be removed
// from the map
func (s *Store) FinishIndexing() {
	err := s.db.WriteBatch(func(wb storage.Batch) error {
		for iter := range s.syncedDocumentURIs.IterBuffered() {
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

func (s *Store) writeAllSymbols(batch storage.Batch, comBatch storage.Batch, document *Document,
	ciDeletor *completionIndexDeletor, syDeletor *symbolDeletor) {
	var documentNamespaces []documentNamespace
	for _, impTable := range document.importTables {
		is, namespaceKey := indexablesFromNamespaceName(impTable.GetNamespace())
		for index, i := range is {
			key := namespaceKey + KeySep + strconv.Itoa(index)
			s.indexName(comBatch, document, i, key)
			ciDeletor.MarkNotDelete(document.GetURI(), i, key)
		}
	}
	tra := newTraverser()
	var referenceEntryInfos []entryInfo
	var documentSymbols []documentSymbol
	tra.traverseDocument(document, func(tra *traverser, child Symbol, _ []Symbol) {
		if ser, ok := child.(serialisable); ok {
			key := ser.GetKey()
			if key == "" {
				return
			}
			entry := newEntry(ser.GetCollection(), key)
			ser.Serialise(entry.e)
			writeEntry(batch, entry)
			symbol := documentSymbol{
				collection: ser.GetCollection(),
				key:        ser.GetKey(),
			}
			syDeletor.MarkNotDelete(ser)

			if indexable, ok := child.(NameIndexable); ok {
				s.indexName(comBatch, document, indexable, key)
				ciDeletor.MarkNotDelete(document.GetURI(), indexable, key)
			}
			documentSymbols = append(documentSymbols, symbol)
		}

		if r, ok := child.(SymbolReference); ok {
			referenceEntryInfos = append(referenceEntryInfos, entryInfo{
				ref: r.ReferenceFQN(),
				r:   r.ReferenceLocation().Range,
			})
		}
		if h, ok := child.(HasTypes); ok {
			location := h.GetLocation()
			if !stub.IsStub(location.URI) {
				r := location.Range
				for _, ref := range SymToRefs(document, h) {
					referenceEntryInfos = append(referenceEntryInfos, entryInfo{
						ref: ref,
						r:   r,
					})
				}
			}
		}
	})
	s.indexDocumentSymbols(document, batch, documentSymbols)
	s.indexDocumentNamespaces(document, batch, documentNamespaces)
	s.refIndex.index(s, document, batch, referenceEntryInfos)
}

func (s *Store) indexDocumentSymbols(document *Document, batch storage.Batch, symbols []documentSymbol) {
	uri := document.GetURI()
	entry := newEntry(documentSymbolsCollection, uri)
	entry.e.WriteInt(len(symbols))
	for _, symbol := range symbols {
		symbol.serialise(entry.e)
	}
	s.greb.Put(entry.getKeyBytes(), entry.e.Bytes())
}

func (s *Store) indexName(batch storage.Batch, document *Document, indexable NameIndexable, key string) {
	entries := createCompletionEntries(document.GetURI(), indexable, key)
	for _, entry := range entries {
		writeEntry(batch, entry)
	}
}

func (s *Store) indexDocumentNamespaces(document *Document, batch storage.Batch, namespaces []documentNamespace) {
	uri := document.GetURI()
	entry := newEntry(documentNamespacesCollection, uri)
	entry.e.WriteInt(len(namespaces))
	for _, ns := range namespaces {
		ns.serialise(entry.e)
	}
	s.greb.Put(entry.getKeyBytes(), entry.e.Bytes())
}

func writeEntry(batch storage.Batch, entry *entry) {
	batch.Put(entry.getKeyBytes(), entry.bytes())
}

// GetURI returns the store URI
func (s *Store) GetURI() protocol.DocumentURI {
	return s.uri
}

// IsSymbolValid returns true if the given symbol matches all predicates of the given options
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

// SearchNamespaces searches namespaces with the given keyword, and keyword can contain
// a namespace scope, e.g. Namespace1\NestedNams
func (s *Store) SearchNamespaces(keyword string, options SearchOptions) ([]string, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
	// In namespace normally there isn't \ but somehow it has ignores in because
	// namespaces are not indexed with \
	if scope == "\\" {
		scope = ""
	}
	namespaces := []string{}
	collection := namespaceCompletionIndex
	if len(scope) != 0 {
		collection += KeySep + scope
	}
	query := searchQuery{
		collection: collection,
		keyword:    keyword,
		onData: func(cv CompletionValue) onDataResult {
			parts := strings.Split(string(cv), KeySep)
			namespaces = append(namespaces, parts[0])
			return onDataResult{false}
		},
	}
	result := searchCompletions(s.comDB, query)
	return namespaces, result
}

// GetClasses searches all classes with the given name
func (s *Store) GetClasses(name string) []*Class {
	entry := newEntry(classCollection, name+KeySep)
	classes := []*Class{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classes = append(classes, ReadClass(d))
	})
	return classes
}

func (s *Store) getClassesByScopeStream(scope string, onData func(*Class) onDataResult) {
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

// SearchClasses uses the completion index to search for classes with the given keyword.
// `keyword` can contain scope \Namespace1\Cl
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
		s.getClassesByScopeStream(scope, func(class *Class) onDataResult {
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
			if len(value) == 0 {
				log.Printf("Empty class: %v", strings.Split(string(completionValue), KeySep))
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
	result := searchCompletions(s.comDB, query)
	return classes, result
}

// GetInterfaces searches all the interfaces with the given name from the disk storage
func (s *Store) GetInterfaces(name string) []*Interface {
	entry := newEntry(interfaceCollection, name+KeySep)
	interfaces := []*Interface{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		interfaces = append(interfaces, ReadInterface(d))
	})
	return interfaces
}

// SearchInterfaces uses completion index to search for interfaces with the given keyword.
// `keyword` can contain scope \Namespace1\Cl
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
	result := searchCompletions(s.comDB, query)
	return interfaces, result
}

// GetTraits searches for all the traits with the given name from the disk storage
func (s *Store) GetTraits(name string) []*Trait {
	entry := newEntry(traitCollection, name+KeySep)
	traits := []*Trait{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		traits = append(traits, ReadTrait(d))
	})
	return traits
}

// SearchTraits uses completion index to search traits matching the given keyword.
// `keyword` can contain scope
func (s *Store) SearchTraits(keyword string, options SearchOptions) ([]*Trait, SearchResult) {
	scope, keyword := GetScopeAndNameFromString(keyword)
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
	result := searchCompletions(s.comDB, query)
	return traits, result
}

// GetFunctions searches all functions with the given name from the disk storage
func (s *Store) GetFunctions(name string) []*Function {
	entry := newEntry(functionCollection, name+KeySep)
	functions := []*Function{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		functions = append(functions, ReadFunction(d))
	})
	return functions
}

// SearchFunctions uses the completion index to search functions matching the given keyword.
// `keyword` can contain scope
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
			if len(value) == 0 {
				log.Printf("Empty function: %v", strings.Split(string(completionValue), KeySep))
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
	result := searchCompletions(s.comDB, query)
	return functions, result
}

// GetConsts searches all the consts with the given name from the disk storage
func (s *Store) GetConsts(name string) []*Const {
	entry := newEntry(constCollection, name+KeySep)
	consts := []*Const{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		consts = append(consts, ReadConst(d))
	})
	return consts
}

// SearchConsts uses completion index to search constants matching the given keyword
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
	result := searchCompletions(s.comDB, query)
	return consts, result
}

// GetDefines searches all `define()` with the given name from the disk storage
func (s *Store) GetDefines(name string) []*Define {
	entry := newEntry(defineCollection, name+KeySep)
	defines := []*Define{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		defines = append(defines, ReadDefine(d))
	})
	return defines
}

// SearchDefines uses completion index to search `define()`s matching the given keyword
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
	result := searchCompletions(s.comDB, query)
	return defines, result
}

// GetMethods searches for all methods with the same scope and name as given
func (s *Store) GetMethods(scope string, name string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep+strings.ToLower(name)+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		methods = append(methods, ReadMethod(d))
	})
	return methods
}

// GetAllMethods returns all the methods with the given scope.
// This function can be faster than `SearchMethods` for searching only methods
// under given scope, because this only scans methods which have the given scope
func (s *Store) GetAllMethods(scope string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep)
	methods := []*Method{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		methods = append(methods, ReadMethod(d))
	})
	return methods
}

// SearchMethods uses completion index to search methods matching the given scope and keyword.
// This function is slow and should only be used for searching method store-wide, because the completion
// index will scan through all the methods in the store and compare its scope.
// If the scope is "" all methods matching will be returned.
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
			if len(value) == 0 {
				log.Printf("Empty methods: scope %s, %v", scope, strings.Split(string(completionValue), KeySep))
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
	result := searchCompletions(s.comDB, query)
	return methods, result
}

// GetClassConsts searches all class constants with the given scope and name from the disk
// storage.
// The word class is used loosely in here which means it can be interfaces/traits
func (s *Store) GetClassConsts(scope string, name string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep+name)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classConsts = append(classConsts, ReadClassConst(d))
	})
	return classConsts
}

// GetAllClassConsts returns all the class constants under the given scope.
// The word class is used loosely in here, which means it can be classes/interfaces/traits
func (s *Store) GetAllClassConsts(scope string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep)
	classConsts := []*ClassConst{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		classConsts = append(classConsts, ReadClassConst(d))
	})
	return classConsts
}

// SearchClassConsts uses completion index to search class constants matching the
// given scope and keyword.
// If the scope is empty all matched class constants are returned.
// The word class is used loosely in here, which means it can be classes/interfaces/traits
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
	result := searchCompletions(s.comDB, query)
	return classConsts, result
}

// GetProperties searches all properties with the given scope and name from the disk storage
func (s *Store) GetProperties(scope string, name string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep+name+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		properties = append(properties, ReadProperty(d))
	})
	return properties
}

// GetAllProperties searches all properties with the given scope from the disk storage
func (s *Store) GetAllProperties(scope string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep)
	properties := []*Property{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		properties = append(properties, ReadProperty(d))
	})
	return properties
}

// SearchProperties uses completion index to search properties matching the given scope
// and name. If the scope is "", this will forward to `GetAllProperties`, and ignore keyword
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
	result := searchCompletions(s.comDB, query)
	return properties, result
}

// GetGlobalVariables searches all global variables with the given name from the disk storage
func (s *Store) GetGlobalVariables(name string) []*GlobalVariable {
	entry := newEntry(globalVariableCollection, name+KeySep)
	results := []*GlobalVariable{}
	s.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		d := storage.NewDecoder(it.Value())
		results = append(results, ReadGlobalVariable(d))
	})
	return results
}

// GetReferences returns the locations of the reference to an FQN
func (s *Store) GetReferences(ref string) []protocol.Location {
	return s.refIndex.search(s, ref)
}

// SymToRefs converts a HasTypes symbol to reference strings
func SymToRefs(document *Document, sym HasTypes) []string {
	var refs []string
	switch v := sym.(type) {
	case *FunctionCall:
		possibleFQNs := document.ImportTableAtPos(sym.GetLocation().Range.Start).FunctionPossibleFQNs(NewTypeString(v.Name))
		for _, fqn := range possibleFQNs {
			sb := &strings.Builder{}
			sb.WriteString(fqn)
			sb.WriteString("()")
			refs = append(refs, sb.String())
		}
	case *ClassTypeDesignator, *TypeDeclaration, *ClassAccess, *TraitAccess, *InterfaceAccess:
		if c, ok := sym.(*ClassAccess); ok && IsNameRelative(c.Name) {
			break
		}
		for _, t := range sym.GetTypes().Resolve() {
			refs = append(refs, t.GetFQN())
		}
	case HasTypesHasScope:
		sb := &strings.Builder{}
		sb.WriteString(".")
		sb.WriteString(v.MemberName())
		refs = append(refs, sb.String())
	}
	return refs
}
