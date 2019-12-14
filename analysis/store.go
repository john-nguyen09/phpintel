package analysis

import (
	"bytes"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	putil "github.com/john-nguyen09/phpintel/util"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
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

const scopeSep = "::"

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

func (s *entry) prefixRange() *util.Range {
	return util.BytesPrefix(s.getKeyBytes())
}

func (s *entry) getKeyBytes() []byte {
	return s.key
}

func (s *entry) getBytes() []byte {
	return s.serialiser.GetBytes()
}

type Store struct {
	uri       protocol.DocumentURI
	db        *leveldb.DB
	documents cmap.ConcurrentMap

	syncedDocumentURIs cmap.ConcurrentMap
}

func NewStore(uri protocol.DocumentURI, storePath string) (*Store, error) {
	db, err := leveldb.OpenFile(storePath, nil)
	if err != nil {
		return nil, err
	}
	return &Store{
		uri:       uri,
		db:        db,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
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
	} else {
		document = value.(*Document)
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
	document.Load()
	s.SyncDocument(document)
}

func (s *Store) CloseDocument(uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("document %s not found", uri)
		return
	}
	document.Close()
	s.SyncDocument(document)
}

func (s *Store) CreateDocument(uri protocol.DocumentURI) {
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		return
	}
	document.Load()
	s.SyncDocument(document)
}

func (s *Store) DeleteDocument(uri protocol.DocumentURI) {
	batch := new(leveldb.Batch)
	s.forgetDocument(batch, uri)
	err := s.db.Write(batch, nil)
	if err != nil {
		log.Println(err)
	}
}

func (s *Store) DeleteFolder(uri protocol.DocumentURI) {
	entry := newEntry(documentCollection, uri)
	iter := s.db.NewIterator(entry.prefixRange(), nil)
	for iter.Next() {
		uri := strings.Split(string(iter.Key()), KeySep)[1]
		s.DeleteDocument(uri)
	}
}

func (s *Store) CompareAndIndexDocument(filePath string) {
	uri := putil.PathToUri(filePath)
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		return
	}

	currentMD5 := document.GetMD5Hash()
	savedMD5, ok := s.syncedDocumentURIs.Get(uri)
	if ok {
		s.syncedDocumentURIs.Remove(uri)
	}
	if ok && bytes.Compare(currentMD5, savedMD5.([]byte)) == 0 {
		return
	}

	document.Load()
	s.SyncDocument(document)
}

func (s *Store) ChangeDocument(uri string, changes []protocol.TextDocumentContentChangeEvent) error {
	defer putil.TimeTrack(time.Now(), "ChangeDocument")
	document := s.GetOrCreateDocument(uri)
	if document == nil {
		log.Printf("Document %s not found", uri)
	}
	document.ApplyChanges(changes)
	s.SyncDocument(document)
	return nil
}

func (s *Store) SyncDocument(document *Document) {
	batch := new(leveldb.Batch)
	s.forgetAllSymbols(batch, document.GetURI())
	s.writeAllSymbols(batch, document)
	entry := newEntry(documentCollection, document.GetURI())
	batch.Put(entry.getKeyBytes(), document.GetMD5Hash())
	err := s.db.Write(batch, nil)
	if err != nil {
		log.Print(err)
	}
	if document.IsOpen() {
		s.documents.Set(document.uri, document)
	} else {
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
	batch := new(leveldb.Batch)
	for iter := range s.syncedDocumentURIs.Iter() {
		s.forgetDocument(batch, iter.Key)
		s.syncedDocumentURIs.Remove(iter.Key)
	}
	err := s.db.Write(batch, nil)
	if err != nil {
		log.Println(err)
	}
}

func (s *Store) getSyncedDocumentURIs() map[string][]byte {
	documentURIs := make(map[string][]byte)
	entry := newEntry(documentCollection, "")
	iterator := s.db.NewIterator(entry.prefixRange(), nil)
	for iterator.Next() {
		key := string(iterator.Key())
		value := iterator.Value()
		value = append(value[:0:0], value...)
		documentURIs[strings.Split(key, KeySep)[1]] = value
	}
	return documentURIs
}

func (s *Store) forgetDocument(batch *leveldb.Batch, uri string) {
	s.forgetAllSymbols(batch, uri)
	entry := newEntry(documentCollection, uri)
	batch.Delete(entry.getKeyBytes())
}

func (s *Store) forgetAllSymbols(batch *leveldb.Batch, uri string) {
	entry := newEntry(documentSymbols, uri+KeySep)
	it := s.db.NewIterator(entry.prefixRange(), nil)
	defer it.Release()
	for it.Next() {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		toBeDelete := newEntry(keyInfo[2], strings.Join(keyInfo[3:], KeySep))
		deleteEntry(batch, toBeDelete)
	}
	deleteCompletionIndex(s.db, batch, uri)
}

func (s *Store) writeAllSymbols(batch *leveldb.Batch, document *Document) {
	for _, child := range document.Children {
		if serialisable, ok := child.(Serialisable); ok {
			key := serialisable.GetKey()
			if key == "" {
				continue
			}
			entry := newEntry(serialisable.GetCollection(), key)
			serialisable.Serialise(entry.serialiser)
			writeEntry(batch, entry)
			rememberSymbol(batch, document, serialisable)

			if indexable, ok := child.(NameIndexable); ok {
				indexName(batch, document, indexable, key)
			}
		}
	}
}

func rememberSymbol(batch *leveldb.Batch, document *Document, serialisable Serialisable) {
	entry := newEntry(documentSymbols, document.GetURI()+KeySep+serialisable.GetCollection()+KeySep+serialisable.GetKey())
	writeEntry(batch, entry)
}

func indexName(batch *leveldb.Batch, document *Document, indexable NameIndexable, key string) {
	entries := createCompletionEntries(document.GetURI(), indexable, key)
	for _, entry := range entries {
		writeEntry(batch, entry)
	}
}

func writeEntry(batch *leveldb.Batch, entry *entry) {
	batch.Put(entry.getKeyBytes(), entry.getBytes())
}

func deleteEntry(batch *leveldb.Batch, entry *entry) {
	batch.Delete(entry.getKeyBytes())
}

func (s *Store) GetURI() protocol.DocumentURI {
	return s.uri
}

func (s *Store) GetClasses(name string) []*Class {
	entry := newEntry(classCollection, name+KeySep)
	classes := []*Class{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		classes = append(classes, ReadClass(serialiser))
	}
	return classes
}

func (s *Store) SearchClasses(keyword string) []*Class {
	completionValues := searchCompletions(s.db, classCompletionIndex, keyword, "")
	classes := []*Class{}
	for _, completionValue := range completionValues {
		entry := newEntry(classCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		classes = append(classes, ReadClass(serialiser))
	}
	return classes
}

func (s *Store) GetInterfaces(name string) []*Interface {
	entry := newEntry(interfaceCollection, name+KeySep)
	interfaces := []*Interface{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		interfaces = append(interfaces, ReadInterface(serialiser))
	}
	return interfaces
}

func (s *Store) SearchInterfaces(keyword string) []*Interface {
	completionValues := searchCompletions(s.db, interfaceCompletionIndex, keyword, "")
	interfaces := []*Interface{}
	for _, completionValue := range completionValues {
		entry := newEntry(interfaceCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		interfaces = append(interfaces, ReadInterface(serialiser))
	}
	return interfaces
}

func (s *Store) GetTraits(name string) []*Trait {
	entry := newEntry(traitCollection, name+KeySep)
	traits := []*Trait{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		traits = append(traits, ReadTrait(serialiser))
	}
	return traits
}

func (s *Store) SearchTraits(keyword string) []*Trait {
	completionValues := searchCompletions(s.db, traitCompletionIndex, keyword, "")
	traits := []*Trait{}
	for _, completionValue := range completionValues {
		entry := newEntry(traitCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		traits = append(traits, ReadTrait(serialiser))
	}
	return traits
}

func (s *Store) GetFunctions(name string) []*Function {
	entry := newEntry(functionCollection, name+KeySep)
	functions := []*Function{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		functions = append(functions, ReadFunction(serialiser))
	}
	return functions
}

func (s *Store) SearchFunctions(keyword string) []*Function {
	completionValues := searchCompletions(s.db, functionCompletionIndex, keyword, "")
	functions := []*Function{}
	for _, completionValue := range completionValues {
		entry := newEntry(functionCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		functions = append(functions, ReadFunction(serialiser))
	}
	return functions
}

func (s *Store) GetConsts(name string) []*Const {
	entry := newEntry(constCollection, name+KeySep)
	consts := []*Const{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		consts = append(consts, ReadConst(serialiser))
	}
	return consts
}

func (s *Store) SearchConsts(keyword string) []*Const {
	completionValues := searchCompletions(s.db, constCompletionIndex, keyword, "")
	consts := []*Const{}
	for _, completionValue := range completionValues {
		entry := newEntry(constCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		consts = append(consts, ReadConst(serialiser))
	}
	return consts
}

func (s *Store) GetDefines(name string) []*Define {
	entry := newEntry(defineCollection, name+KeySep)
	defines := []*Define{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		defines = append(defines, ReadDefine(serialiser))
	}
	return defines
}

func (s *Store) SearchDefines(keyword string) []*Define {
	completionValues := searchCompletions(s.db, defineCompletionIndex, keyword, "")
	defines := []*Define{}
	for _, completionValue := range completionValues {
		entry := newEntry(defineCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		defines = append(defines, ReadDefine(serialiser))
	}
	return defines
}

func (s *Store) GetMethods(scope string, name string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep+name+KeySep)
	methods := []*Method{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		methods = append(methods, ReadMethod(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		methods = append(methods, class.GetInheritedMethods(s, name, methods)...)
	}
	return methods
}

func (s *Store) GetAllMethods(scope string) []*Method {
	entry := newEntry(methodCollection, scope+KeySep)
	methods := []*Method{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		methods = append(methods, ReadMethod(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		methods = append(methods, class.SearchInheritedMethods(s, "", methods)...)
	}
	return methods
}

func (s *Store) SearchMethods(scope string, keyword string) []*Method {
	if keyword == "" {
		return s.GetAllMethods(scope)
	}

	completionValues := searchCompletions(s.db, methodCompletionIndex, keyword, scope)
	methods := []*Method{}
	for _, completionValue := range completionValues {
		entry := newEntry(methodCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		methods = append(methods, ReadMethod(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		methods = append(methods, class.SearchInheritedMethods(s, keyword, methods)...)
	}
	return methods
}

func (s *Store) GetClassConsts(scope string, name string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep+name)
	classConsts := []*ClassConst{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		classConsts = append(classConsts, ReadClassConst(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		classConsts = append(classConsts, class.GetInheritedClassConsts(s, name)...)
	}
	return classConsts
}

func (s *Store) GetAllClassConsts(scope string) []*ClassConst {
	entry := newEntry(classConstCollection, scope+KeySep)
	classConsts := []*ClassConst{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		classConsts = append(classConsts, ReadClassConst(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		classConsts = append(classConsts, class.SearchInheritedClassConsts(s, "")...)
	}
	return classConsts
}

func (s *Store) SearchClassConsts(scope string, keyword string) []*ClassConst {
	if keyword == "" {
		s.GetAllClassConsts(scope)
	}

	completionValues := searchCompletions(s.db, classConstCompletionIndex, keyword, scope)
	classConsts := []*ClassConst{}
	for _, completionValue := range completionValues {
		entry := newEntry(classConstCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		classConsts = append(classConsts, ReadClassConst(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		classConsts = append(classConsts, class.SearchInheritedClassConsts(s, keyword)...)
	}
	return classConsts
}

func (s *Store) GetProperties(scope string, name string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep+name)
	properties := []*Property{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		properties = append(properties, ReadProperty(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		properties = append(properties, class.GetInheritedProperties(s, name, properties)...)
	}
	return properties
}

func (s *Store) GetAllProperties(scope string) []*Property {
	entry := newEntry(propertyCollection, scope+KeySep)
	properties := []*Property{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		properties = append(properties, ReadProperty(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		properties = append(properties, class.SearchInheritedProperties(s, "", properties)...)
	}
	return properties
}

func (s *Store) SearchProperties(scope string, keyword string) []*Property {
	if keyword == "" {
		return s.GetAllProperties(scope)
	}

	completionValues := searchCompletions(s.db, propertyCompletionIndex, keyword, scope)
	properties := []*Property{}
	for _, completionValue := range completionValues {
		entry := newEntry(propertyCollection, string(completionValue))
		theBytes, err := s.db.Get(entry.getKeyBytes(), nil)
		if err != nil {
			continue
		}
		serialiser := SerialiserFromByteSlice(theBytes)
		properties = append(properties, ReadProperty(serialiser))
	}
	classes := s.GetClasses(scope)
	for _, class := range classes {
		properties = append(properties, class.SearchInheritedProperties(s, keyword, properties)...)
	}
	return properties
}

func (s *Store) GetGlobalVariables(name string) []*GlobalVariable {
	entry := newEntry(globalVariableCollection, name+KeySep)
	results := []*GlobalVariable{}
	it := s.db.NewIterator(entry.prefixRange(), nil)
	for it.Next() {
		serialiser := SerialiserFromByteSlice(it.Value())
		results = append(results, ReadGlobalVariable(serialiser))
	}
	return results
}
