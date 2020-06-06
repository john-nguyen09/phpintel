package analysis

import (
	"strings"
)

const sep = ":"
const staticInAccessCost = 16

// Query is a wrapper around store
type Query struct {
	store *Store
	cache map[string]interface{}
}

// NewQuery creates a new query, a query should not outlive a request
func NewQuery(store *Store) *Query {
	return &Query{
		store: store,
		cache: make(map[string]interface{}),
	}
}

// Store returns the store of the query
func (q Query) Store() *Store {
	return q.store
}

// GetClasses is a cached proxy behind store
func (q *Query) GetClasses(name string) []*Class {
	cacheKey := "Classes" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if classes, ok := data.([]*Class); ok {
			return classes
		}
	}
	classes := q.store.GetClasses(name)
	q.cache[cacheKey] = classes
	return classes
}

// GetClassConstructor returns the constructor of the given class
func (q *Query) GetClassConstructor(class *Class) MethodWithScope {
	methods := q.GetMethods(class.Name.GetFQN(), "__construct")
	result := MethodWithScope{}
	if len(methods) > 0 {
		result.Method = methods[0]
		result.Scope = class
	}
	return result
}

// GetInterfaces is a cached proxy behind store
func (q *Query) GetInterfaces(name string) []*Interface {
	cacheKey := "Interfaces" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if interfaces, ok := data.([]*Interface); ok {
			return interfaces
		}
	}
	interfaces := q.store.GetInterfaces(name)
	q.cache[cacheKey] = interfaces
	return interfaces
}

// GetTraits is a cached proxy behind store
func (q *Query) GetTraits(name string) []*Trait {
	cacheKey := "Traits" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if traits, ok := data.([]*Trait); ok {
			return traits
		}
	}
	traits := q.store.GetTraits(name)
	q.cache[cacheKey] = traits
	return traits
}

// GetFunctions is a cached proxy behind store
func (q *Query) GetFunctions(name string) []*Function {
	cacheKey := "Functions" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if functions, ok := data.([]*Function); ok {
			return functions
		}
	}
	functions := q.store.GetFunctions(name)
	q.cache[cacheKey] = functions
	return functions
}

// GetConsts is a cached proxy behind store
func (q *Query) GetConsts(name string) []*Const {
	cacheKey := "Consts" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if consts, ok := data.([]*Const); ok {
			return consts
		}
	}
	consts := q.store.GetConsts(name)
	q.cache[cacheKey] = consts
	return consts
}

// GetDefines is a cached proxy behind store
func (q *Query) GetDefines(name string) []*Define {
	cacheKey := "Defines" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if defines, ok := data.([]*Define); ok {
			return defines
		}
	}
	defines := q.store.GetDefines(name)
	q.cache[cacheKey] = defines
	return defines
}

// GetGlobalVariables returns global variables with the given name
func (q *Query) GetGlobalVariables(name string) []*GlobalVariable {
	cacheKey := "GlobalVariables" + sep + name
	if data, ok := q.cache[cacheKey]; ok {
		if v, ok := data.([]*GlobalVariable); ok {
			return v
		}
	}
	v := q.store.GetGlobalVariables(name)
	q.cache[cacheKey] = v
	return v
}

// MethodWithScope represents a method with its scope
type MethodWithScope struct {
	Method *Method
	Scope  Symbol
	Score  int
}

func methodWithScopeFromMethods(scope Symbol, methods []*Method) []MethodWithScope {
	results := []MethodWithScope{}
	for _, method := range methods {
		results = append(results, MethodWithScope{method, scope, 0})
	}
	return results
}

// MergeMethodWithScope returns a merged methods with scope
func MergeMethodWithScope(items ...[]MethodWithScope) []MethodWithScope {
	results := []MethodWithScope{}
	duplicated := map[string]struct{}{}
	for _, methods := range items {
		for _, method := range methods {
			key := method.Method.Name
			if s, ok := method.Scope.(serialisable); ok {
				key = s.GetKey() + "::" + method.Method.Name
			}
			if _, ok := duplicated[key]; ok {
				continue
			}
			results = append(results, method)
			duplicated[key] = struct{}{}
		}
	}
	return results
}

// Relations represents the related FQNs
type Relations map[string]struct{}

// Relate sets the related FQNs
func (r Relations) Relate(otherFQN string) {
	r[otherFQN] = struct{}{}
}

// IsRelated checks if the other FQN is related
func (r Relations) IsRelated(otherFQN string) bool {
	if _, ok := r[otherFQN]; ok {
		return true
	}
	return false
}

// RelationMap is the map for relationships
type RelationMap map[string]Relations

// IsRelated checks if this FQN is related to other FQN
func (r RelationMap) IsRelated(thisFQN string, otherFQN string) bool {
	if relations, ok := r[thisFQN]; ok && relations.IsRelated(otherFQN) {
		return true
	}
	return false
}

// Relate sets the related FQNs
func (r RelationMap) Relate(thisFQN string, otherFQN string) {
	var curr Relations
	if m, ok := r[thisFQN]; ok {
		curr = m
	} else {
		curr = Relations{}
	}
	curr.Relate(otherFQN)
	r[thisFQN] = curr
}

// Merge merges other relation map to current
func (r RelationMap) Merge(other RelationMap) {
	for fqn, relations := range other {
		for otherFQN := range relations {
			r.Relate(fqn, otherFQN)
		}
	}
}

// IsInheritedStatic checks the given conditions if it can be inherited
func IsInheritedStatic(currentClass string, access MemberAccess, relationMap RelationMap, member MemberSymbol) bool {
	scopeName := access.ScopeName()
	scopeTypes := access.ScopeTypes()
	memberScopeFQN := member.ScopeTypeString().GetFQN()
	visibility := member.Visibility()
	isStatic := member.IsStatic()
	if IsNameParent(scopeName) {
		return memberScopeFQN != currentClass && visibility != Private
	}
	if IsNameRelative(scopeName) {
		return memberScopeFQN == currentClass || visibility != Private
	}
	if !isStatic {
		return false
	}
	if visibility == Public {
		return true
	}
	var (
		isSameClass bool
		isRelated   bool
	)
	for _, ts := range scopeTypes.Resolve() {
		fqn := ts.GetFQN()
		if fqn == memberScopeFQN {
			isSameClass = true
			break
		}
		if relationMap.IsRelated(memberScopeFQN, fqn) {
			isRelated = true
			break
		}
	}
	return isSameClass || (isRelated && visibility == Protected)
}

// IsInherited checks whether the given conditions can inherit the symbol with non-static rules
func IsInherited(currentClass string, access MemberAccess, relationMap RelationMap, member MemberSymbol) bool {
	scopeName := access.ScopeName()
	scopeTypes := access.ScopeTypes()
	memberScopeFQN := member.ScopeTypeString().GetFQN()
	visibility := member.Visibility()
	if scopeName == "$this" && (memberScopeFQN == currentClass || visibility != Private) {
		return true
	}
	if visibility == Public {
		return true
	}
	var (
		isSameClass bool
		isRelated   bool
	)
	for _, ts := range scopeTypes.Resolve() {
		fqn := ts.GetFQN()
		if fqn == currentClass {
			isSameClass = true
			break
		}
		if relationMap.IsRelated(memberScopeFQN, fqn) {
			isRelated = true
			break
		}
	}
	return isSameClass || (isRelated && visibility == Protected)
}

// InheritedMethods contains the methods and the searched scope names
type InheritedMethods struct {
	Methods      []MethodWithScope
	RelationMap  RelationMap
	SearchedFQNs map[string]struct{}
}

// EmptyInheritedMethods returns an empty inherited methods
func EmptyInheritedMethods() InheritedMethods {
	return NewInheritedMethods(nil, make(map[string]struct{}))
}

// NewInheritedMethods returns InheritedMethods struct
func NewInheritedMethods(methods []MethodWithScope, searchedFQNs map[string]struct{}) InheritedMethods {
	return InheritedMethods{
		Methods:      methods,
		RelationMap:  RelationMap{},
		SearchedFQNs: searchedFQNs,
	}
}

// Merge merges current inherited methods with others, this merges
// the underlying methods and searched FQNs
func (m *InheritedMethods) Merge(other InheritedMethods) {
	m.Methods = append(m.Methods, other.Methods...)
	if m.SearchedFQNs == nil {
		m.SearchedFQNs = make(map[string]struct{})
	}
	for fqn := range other.SearchedFQNs {
		m.SearchedFQNs[fqn] = struct{}{}
	}
	m.RelationMap.Merge(other.RelationMap)
}

// ReduceStatic filters the methods by the static rules, even though
// the methods are not necessarily static, e.g. self::NonStaticMethod().
// Rules:
// - If the `name` is parent, includes methods not from current class and not private
// - If the `name` is relative (static, self), includes methods from same class or not
//   private methods
// - Otherwise, includes methods that are static and public
func (m InheritedMethods) ReduceStatic(currentClass string, access MemberAccess) []MethodWithScope {
	var results []MethodWithScope
	duplicatedNames := make(map[string]struct{})
	for _, ms := range m.Methods {
		if _, ok := duplicatedNames[ms.Method.Name]; ok {
			continue
		}
		if IsInheritedStatic(currentClass, access, m.RelationMap, ms.Method) {
			results = append(results, ms)
			duplicatedNames[ms.Method.Name] = struct{}{}
		}
	}
	return results
}

// ReduceAccess filters the methods by the access rules
// Rules:
// - If the `name` is $this includes the scope is the current class or not private methods
// - If the type of `scope` is the same as the current class, private methods can
//   be accessed
// - Else, only public methods can be accessed
func (m InheritedMethods) ReduceAccess(currentClass string, access MemberAccess) []MethodWithScope {
	var results []MethodWithScope
	duplicatedNames := make(map[string]struct{})
	for _, ms := range m.Methods {
		if _, ok := duplicatedNames[ms.Method.Name]; ok {
			continue
		}
		if IsInherited(currentClass, access, m.RelationMap, ms.Method) {
			if ms.Method.isStatic {
				ms.Score -= staticInAccessCost
			}
			results = append(results, ms)
			duplicatedNames[ms.Method.Name] = struct{}{}
		}
	}
	return results
}

// Len returns number of methods
func (m InheritedMethods) Len() int {
	return len(m.Methods)
}

// GetMethods searches for all methods under the given scope, this function
// does not consider inheritance, if name is empty this will return all methods
func (q *Query) GetMethods(scope string, name string) []*Method {
	cacheKey := "Methods" + sep + scope + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if methods, ok := data.([]*Method); ok {
			return methods
		}
	}
	var methods []*Method
	if name != "" {
		methods = q.store.GetMethods(scope, name)
	} else {
		methods = q.store.GetAllMethods(scope)
	}
	q.cache[cacheKey] = methods
	return methods
}

// GetClassMethods returns all methods from the given class, its extends and implements
func (q *Query) GetClassMethods(class *Class, name string, searchedFQNs map[string]struct{}) InheritedMethods {
	classes := []*Class{
		class,
	}
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	cacheKey := "ClassMethods" + sep + class.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if methods, ok := data.(InheritedMethods); ok {
			return methods
		}
	}
	methods := NewInheritedMethods(nil, searchedFQNs)
	methodsFromInterfaces := NewInheritedMethods(nil, searchedFQNs)
	for len(classes) > 0 {
		var class *Class
		class, classes = classes[0], classes[1:]
		scope := class.Name.GetFQN()
		if _, ok := searchedFQNs[scope]; ok {
			continue
		}
		searchedFQNs[scope] = struct{}{}
		methods.Methods = append(methods.Methods, methodWithScopeFromMethods(class, q.GetMethods(scope, name))...)
		for _, use := range class.Use {
			if use.IsEmpty() {
				continue
			}
			for _, trait := range q.store.GetTraits(use.GetFQN()) {
				methods.Merge(q.GetTraitMethods(trait, name))
			}
		}
		if !class.Extends.IsEmpty() {
			if _, ok := searchedFQNs[class.Extends.GetFQN()]; !ok {
				classes = append(classes, q.GetClasses(class.Extends.GetFQN())...)
			}
		}
		for _, implement := range class.Interfaces {
			if implement.IsEmpty() {
				continue
			}
			for _, intf := range q.store.GetInterfaces(implement.GetFQN()) {
				methodsFromInterfaces.Merge(q.GetInterfaceMethods(intf, name, methodsFromInterfaces.SearchedFQNs))
			}
		}
	}
	methods.Merge(methodsFromInterfaces)
	q.cache[cacheKey] = methods
	return methods
}

// GetInterfaceMethods returns all methods from the given interface and its extends
func (q *Query) GetInterfaceMethods(intf *Interface, name string, searchedFQNs map[string]struct{}) InheritedMethods {
	interfaces := []*Interface{
		intf,
	}
	if searchedFQNs == nil {
		searchedFQNs = make(map[string]struct{})
	}
	cacheKey := "InterfaceMethods" + sep + intf.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if methods, ok := data.(InheritedMethods); ok {
			return methods
		}
	}
	methods := NewInheritedMethods(nil, searchedFQNs)
	for len(interfaces) > 0 {
		var intf *Interface
		intf, interfaces = interfaces[0], interfaces[1:]
		scope := intf.Name.GetFQN()
		if _, ok := searchedFQNs[scope]; ok {
			continue
		}
		searchedFQNs[scope] = struct{}{}
		methods.Methods = append(methods.Methods, methodWithScopeFromMethods(intf, q.GetMethods(scope, name))...)
		for _, extend := range intf.Extends {
			if extend.IsEmpty() {
				continue
			}
			for _, intf := range q.store.GetInterfaces(extend.GetFQN()) {
				if _, ok := searchedFQNs[intf.Name.GetFQN()]; !ok {
					interfaces = append(interfaces, intf)
				}
			}
		}
	}
	q.cache[cacheKey] = methods
	return methods
}

// GetTraitMethods returns all methods from the given trait and name
func (q *Query) GetTraitMethods(trait *Trait, name string) InheritedMethods {
	cacheKey := "TraitMethods" + sep + trait.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if methods, ok := data.(InheritedMethods); ok {
			return methods
		}
	}
	methods := NewInheritedMethods(
		methodWithScopeFromMethods(trait, q.GetMethods(trait.Name.GetFQN(), name)),
		make(map[string]struct{}))
	q.cache[cacheKey] = methods
	return methods
}

// SearchClassMethods searches methods using fuzzy
func (q *Query) SearchClassMethods(class *Class, keyword string, searchedFQNs map[string]struct{}) InheritedMethods {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	methods := q.GetClassMethods(class, "", searchedFQNs)
	if keyword != "" {
		newMethods := methods.Methods[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, method := range methods.Methods {
			if matched, score := isMatch(method.Method.Name, patternRune); matched {
				method.Score = score
				newMethods = append(newMethods, method)
			}
		}
		methods.Methods = newMethods
	}
	return methods
}

// SearchInterfaceMethods searches methods using fuzzy
func (q *Query) SearchInterfaceMethods(intf *Interface, keyword string, searchedFQNs map[string]struct{}) InheritedMethods {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	methods := q.GetInterfaceMethods(intf, "", searchedFQNs)
	if keyword != "" {
		newMethods := methods.Methods[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, method := range methods.Methods {
			if matched, score := isMatch(method.Method.Name, patternRune); matched {
				method.Score = score
				newMethods = append(newMethods, method)
			}
		}
		methods.Methods = newMethods
	}
	return methods
}

// ClassConstWithScope is the class const with its scope
type ClassConstWithScope struct {
	Const *ClassConst
	Scope Symbol
	Score int
}

func classConstWithScopeFromConsts(scope Symbol, classConsts []*ClassConst) []ClassConstWithScope {
	results := []ClassConstWithScope{}
	for _, classConst := range classConsts {
		results = append(results, ClassConstWithScope{classConst, scope, 0})
	}
	return results
}

// MergeClassConstWithScope merges the items and remove duplicates
func MergeClassConstWithScope(items ...[]ClassConstWithScope) []ClassConstWithScope {
	results := []ClassConstWithScope{}
	duplicated := map[string]struct{}{}
	for _, classConsts := range items {
		for _, classConst := range classConsts {
			key := classConst.Const.Name
			if s, ok := classConst.Scope.(serialisable); ok {
				key = s.GetKey() + "::" + classConst.Const.Name
			}
			if _, ok := duplicated[key]; ok {
				continue
			}
			results = append(results, classConst)
			duplicated[key] = struct{}{}
		}
	}
	return results
}

// InheritedClassConst contains class consts including inherited ones
type InheritedClassConst struct {
	Consts       []ClassConstWithScope
	RelationMap  RelationMap
	SearchedFQNs map[string]struct{}
}

// EmptyInheritedClassConst creates an empty inherited class const
func EmptyInheritedClassConst() InheritedClassConst {
	return NewInheritedClassConst(nil, make(map[string]struct{}))
}

// NewInheritedClassConst creates inherited
func NewInheritedClassConst(consts []ClassConstWithScope, searchedFQNs map[string]struct{}) InheritedClassConst {
	return InheritedClassConst{
		Consts:       consts,
		RelationMap:  make(map[string]Relations),
		SearchedFQNs: searchedFQNs,
	}
}

// Merge merges current inherited class consts with others
func (c *InheritedClassConst) Merge(other InheritedClassConst) {
	c.Consts = append(c.Consts, other.Consts...)
	for fqn := range other.SearchedFQNs {
		c.SearchedFQNs[fqn] = struct{}{}
	}
	c.RelationMap.Merge(other.RelationMap)
}

// ReduceStatic reduces the inherited props using the static rules
func (c InheritedClassConst) ReduceStatic(currentClass string, access MemberAccess) []ClassConstWithScope {
	results := []ClassConstWithScope{}
	duplicatedNames := make(map[string]struct{})
	for _, cc := range c.Consts {
		if _, ok := duplicatedNames[cc.Const.Name]; ok {
			continue
		}
		if IsInheritedStatic(currentClass, access, c.RelationMap, cc.Const) {
			results = append(results, cc)
			duplicatedNames[cc.Const.Name] = struct{}{}
		}
	}
	return results
}

// GetClassConsts is a cached proxy to store
func (q *Query) GetClassConsts(scope string, name string) []*ClassConst {
	cacheKey := "ClassConst" + sep + scope + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if classConsts, ok := data.([]*ClassConst); ok {
			return classConsts
		}
	}
	var classConsts []*ClassConst
	if name != "" {
		classConsts = q.store.GetClassConsts(scope, name)
	} else {
		classConsts = q.store.GetAllClassConsts(scope)
	}
	q.cache[cacheKey] = classConsts
	return classConsts
}

// GetClassClassConsts returns the inherited class const
func (q *Query) GetClassClassConsts(class *Class, name string, searchedFQNs map[string]struct{}) InheritedClassConst {
	cacheKey := "ClassClassConst" + sep + class.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if classConsts, ok := data.(InheritedClassConst); ok {
			return classConsts
		}
	}
	classes := []*Class{
		class,
	}
	if searchedFQNs == nil {
		searchedFQNs = make(map[string]struct{})
	}
	classConsts := NewInheritedClassConst(nil, searchedFQNs)
	classConstsFromInterfaces := NewInheritedClassConst(nil, searchedFQNs)
	for len(classes) > 0 {
		var class *Class
		class, classes = classes[0], classes[1:]
		classConsts.Consts = append(classConsts.Consts,
			classConstWithScopeFromConsts(class, q.GetClassConsts(class.Name.GetFQN(), name))...)
		if !class.Extends.IsEmpty() {
			if _, ok := searchedFQNs[class.Extends.GetFQN()]; !ok {
				classConsts.RelationMap.Relate(class.Name.GetFQN(), class.Extends.GetFQN())
				classes = append(classes, q.GetClasses(class.Extends.GetFQN())...)
			}
		}
		for _, typeString := range class.Interfaces {
			if typeString.IsEmpty() {
				continue
			}
			for _, intf := range q.GetInterfaces(typeString.GetFQN()) {
				classConsts.RelationMap.Relate(class.Name.GetFQN(), intf.Name.GetFQN())
				classConstsFromInterfaces.Merge(q.GetInterfaceClassConsts(intf, name, classConstsFromInterfaces.SearchedFQNs))
			}
		}
	}
	classConsts.Merge(classConstsFromInterfaces)
	q.cache[cacheKey] = classConsts
	return classConsts
}

// GetInterfaceClassConsts returns the class consts of the interface
func (q *Query) GetInterfaceClassConsts(intf *Interface, name string, searchedFQNs map[string]struct{}) InheritedClassConst {
	interfaces := []*Interface{
		intf,
	}
	if searchedFQNs == nil {
		searchedFQNs = make(map[string]struct{})
	}
	cacheKey := "InterfaceClassConst" + sep + intf.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if cc, ok := data.(InheritedClassConst); ok {
			return cc
		}
	}
	classConsts := NewInheritedClassConst(nil, searchedFQNs)
	for len(interfaces) > 0 {
		intf, interfaces = interfaces[0], interfaces[1:]
		scope := intf.Name.GetFQN()
		if _, ok := searchedFQNs[scope]; ok {
			continue
		}
		searchedFQNs[scope] = struct{}{}
		classConsts.Consts = append(classConsts.Consts,
			classConstWithScopeFromConsts(intf, q.GetClassConsts(scope, name))...)
		for _, extend := range intf.Extends {
			if extend.IsEmpty() {
				continue
			}
			for _, intf := range q.GetInterfaces(extend.GetFQN()) {
				classConsts.RelationMap.Relate(scope, intf.Name.GetFQN())
				if _, ok := searchedFQNs[intf.Name.GetFQN()]; !ok {
					interfaces = append(interfaces, intf)
				}
			}
		}
	}
	q.cache[cacheKey] = classConsts
	return classConsts
}

// SearchClassClassConsts searches for class consts using fuzzy match
func (q *Query) SearchClassClassConsts(class *Class, keyword string, searchedFQNs map[string]struct{}) InheritedClassConst {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	classConsts := q.GetClassClassConsts(class, "", searchedFQNs)
	if keyword != "" {
		newMethods := classConsts.Consts[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, classConst := range classConsts.Consts {
			if matched, score := isMatch(classConst.Const.Name, patternRune); matched {
				classConst.Score = score
				newMethods = append(newMethods, classConst)
			}
		}
		classConsts.Consts = newMethods
	}
	return classConsts
}

// SearchInterfaceClassConsts searches for class consts using fuzzy match
func (q *Query) SearchInterfaceClassConsts(intf *Interface, keyword string, searchedFQNs map[string]struct{}) InheritedClassConst {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	classConsts := q.GetInterfaceClassConsts(intf, "", searchedFQNs)
	if keyword != "" {
		newMethods := classConsts.Consts[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, classConst := range classConsts.Consts {
			if matched, score := isMatch(classConst.Const.Name, patternRune); matched {
				classConst.Score = score
				newMethods = append(newMethods, classConst)
			}
		}
		classConsts.Consts = newMethods
	}
	return classConsts
}

// PropWithScope represents a property with a scope
type PropWithScope struct {
	Prop  *Property
	Scope Symbol
	Score int
}

func propWithScopeFromProps(scope Symbol, props []*Property) []PropWithScope {
	results := []PropWithScope{}
	for _, prop := range props {
		results = append(results, PropWithScope{prop, scope, 0})
	}
	return results
}

// MergePropWithScope returns a merged props with scope
func MergePropWithScope(items ...[]PropWithScope) []PropWithScope {
	results := []PropWithScope{}
	duplicated := map[string]struct{}{}
	for _, props := range items {
		for _, prop := range props {
			key := prop.Prop.Name
			if s, ok := prop.Scope.(serialisable); ok {
				key = s.GetKey() + "::" + prop.Prop.Name
			}
			if _, ok := duplicated[key]; ok {
				continue
			}
			results = append(results, prop)
			duplicated[key] = struct{}{}
		}
	}
	return results
}

// InheritedProps contains information for props include inheried ones
type InheritedProps struct {
	Props        []PropWithScope
	RelationMap  RelationMap
	SearchedFQNs map[string]struct{}
}

// EmptyInheritedProps creates an empty InheritedProps
func EmptyInheritedProps() InheritedProps {
	return NewInheritedProps(nil, make(map[string]struct{}))
}

// NewInheritedProps creates InheritedProps
func NewInheritedProps(props []PropWithScope, searchedFQNs map[string]struct{}) InheritedProps {
	return InheritedProps{
		Props:        props,
		RelationMap:  RelationMap{},
		SearchedFQNs: searchedFQNs,
	}
}

// Merge merges the current inherited props with others
func (p *InheritedProps) Merge(other InheritedProps) {
	p.Props = append(p.Props, other.Props...)
	for fqn := range other.SearchedFQNs {
		p.SearchedFQNs[fqn] = struct{}{}
	}
	p.RelationMap.Merge(other.RelationMap)
}

// ReduceStatic reduces properties using the static rules
func (p InheritedProps) ReduceStatic(currentClass string, access MemberAccess) []PropWithScope {
	results := []PropWithScope{}
	duplicatedNames := make(map[string]struct{})
	for _, ps := range p.Props {
		if _, ok := duplicatedNames[ps.Prop.Name]; ok {
			continue
		}
		if IsInheritedStatic(currentClass, access, p.RelationMap, ps.Prop) {
			results = append(results, ps)
			duplicatedNames[ps.Prop.Name] = struct{}{}
		}
	}
	return results
}

// ReduceAccess reduces propties using the access rules
func (p InheritedProps) ReduceAccess(currentClass string, access MemberAccess) []PropWithScope {
	results := []PropWithScope{}
	duplicatedNames := make(map[string]struct{})
	for _, ps := range p.Props {
		if _, ok := duplicatedNames[ps.Prop.Name]; ok {
			continue
		}
		if IsInherited(currentClass, access, p.RelationMap, ps.Prop) {
			if ps.Prop.isStatic {
				ps.Score -= staticInAccessCost
			}
			results = append(results, ps)
			duplicatedNames[ps.Prop.Name] = struct{}{}
		}
	}
	return results
}

// Len returns number of inherited props
func (p InheritedProps) Len() int {
	return len(p.Props)
}

// GetProps is a cached proxy to store
func (q *Query) GetProps(scope, name string) []*Property {
	cacheKey := "Props" + sep + scope + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if p, ok := data.([]*Property); ok {
			return p
		}
	}
	var p []*Property
	if name != "" {
		p = q.store.GetProperties(scope, name)
	} else {
		p = q.store.GetAllProperties(scope)
	}
	q.cache[cacheKey] = p
	return p
}

// GetClassProps gets all properties for a class
func (q *Query) GetClassProps(class *Class, name string, searchedFQNs map[string]struct{}) InheritedProps {
	classes := []*Class{
		class,
	}
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	cacheKey := "ClassProps" + sep + class.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if p, ok := data.(InheritedProps); ok {
			return p
		}
	}
	props := NewInheritedProps(nil, searchedFQNs)
	propsFromInterfaces := NewInheritedProps(nil, searchedFQNs)
	for len(classes) > 0 {
		class, classes = classes[0], classes[1:]
		props.Props = append(props.Props, propWithScopeFromProps(class, q.GetProps(class.Name.GetFQN(), name))...)
		if !class.Extends.IsEmpty() {
			if _, ok := searchedFQNs[class.Extends.GetFQN()]; !ok {
				classes = append(classes, q.GetClasses(class.Extends.GetFQN())...)
			}
		}
		for _, typeString := range class.Interfaces {
			if typeString.IsEmpty() {
				continue
			}
			for _, intf := range q.GetInterfaces(typeString.GetFQN()) {
				propsFromInterfaces.Merge(q.GetInterfaceProps(intf, name, propsFromInterfaces.SearchedFQNs))
			}
		}
	}
	props.Merge(propsFromInterfaces)
	q.cache[cacheKey] = props
	return props
}

// GetInterfaceProps gets all properties for an interface
func (q *Query) GetInterfaceProps(intf *Interface, name string, searchedFQNs map[string]struct{}) InheritedProps {
	interfaces := []*Interface{
		intf,
	}
	if searchedFQNs == nil {
		searchedFQNs = make(map[string]struct{})
	}
	cacheKey := "InterfaceProps" + sep + intf.Name.GetFQN() + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if p, ok := data.(InheritedProps); ok {
			return p
		}
	}
	props := NewInheritedProps(nil, searchedFQNs)
	for len(interfaces) > 0 {
		intf, interfaces = interfaces[0], interfaces[1:]
		scope := intf.Name.GetFQN()
		if _, ok := searchedFQNs[scope]; ok {
			continue
		}
		searchedFQNs[scope] = struct{}{}
		props.Props = append(props.Props, propWithScopeFromProps(intf, q.GetProps(scope, name))...)
		for _, extend := range intf.Extends {
			if extend.IsEmpty() {
				continue
			}
			for _, intf := range q.GetInterfaces(extend.GetFQN()) {
				if _, ok := searchedFQNs[intf.Name.GetFQN()]; !ok {
					interfaces = append(interfaces, intf)
				}
			}
		}
	}
	q.cache[cacheKey] = props
	return props
}

// SearchClassProps searches for class props using fuzzy match
func (q *Query) SearchClassProps(class *Class, keyword string, searchedFQNs map[string]struct{}) InheritedProps {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	props := q.GetClassProps(class, "", searchedFQNs)
	if keyword != "" {
		newProps := props.Props[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, prop := range props.Props {
			if matched, score := isMatch(prop.Prop.Name, patternRune); matched {
				prop.Score = score
				newProps = append(newProps, prop)
			}
		}
		props.Props = newProps
	}
	return props
}

// SearchInterfaceProps searches for class props using fuzzy match
func (q *Query) SearchInterfaceProps(intf *Interface, keyword string, searchedFQNs map[string]struct{}) InheritedProps {
	if searchedFQNs == nil {
		searchedFQNs = map[string]struct{}{}
	}
	props := q.GetInterfaceProps(intf, "", searchedFQNs)
	if keyword != "" {
		newProps := props.Props[:0]
		patternRune := []rune(strings.ToLower(keyword))
		for _, prop := range props.Props {
			if matched, score := isMatch(prop.Prop.Name, patternRune); matched {
				prop.Score = score
				newProps = append(newProps, prop)
			}
		}
		props.Props = newProps
	}
	return props
}
