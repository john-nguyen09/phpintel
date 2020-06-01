package analysis

const sep = ":"

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
}

func methodWithScopeFromMethods(scope Symbol, methods []*Method) []MethodWithScope {
	results := []MethodWithScope{}
	for _, method := range methods {
		results = append(results, MethodWithScope{method, scope})
	}
	return results
}

// InheritedMethods contains the methods and the searched scope names
type InheritedMethods struct {
	// Methods is the resulted methods
	Methods []MethodWithScope
	// SearchedFQNs is the searched scope names
	SearchedFQNs map[string]struct{}
}

// EmptyInheritedMethods returns an empty inherited methods
func EmptyInheritedMethods() InheritedMethods {
	return InheritedMethods{
		SearchedFQNs: make(map[string]struct{}),
	}
}

// NewInheritedMethods returns InheritedMethods struct
func NewInheritedMethods(methods []MethodWithScope, searchedFQNs map[string]struct{}) InheritedMethods {
	return InheritedMethods{
		Methods:      methods,
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
}

// Reduce returns the filtered out methods based on the given search options
func (m InheritedMethods) Reduce(opts SearchOptions) []*Method {
	var results []*Method
	for _, m := range m.Methods {
		if isSymbolValid(m.Method, opts) {
			results = append(results, m.Method)
		}
	}
	return results
}

// ReduceInherited reduces the method list by their order of occurrence
// this is useful for the results of get*Methods functions because the
// results are in the correct order of inheritance and overriding
func (m InheritedMethods) ReduceInherited() []*Method {
	return m.Reduce(withNoDuplicateNamesOptions(NewSearchOptions()))
}

// ReduceStatic filters the methods by the static rules, even though
// the methods are not necessarily static, e.g. self::NonStaticMethod().
// Rules:
// - If the `name` is parent, includes methods not from current class and not private
// - If the `name` is relative (static, self), includes methods from same class or not
//   private methods
// - Otherwise, includes methods that are static and public
func (m InheritedMethods) ReduceStatic(currentClass, scopeName string) []MethodWithScope {
	var results []MethodWithScope
	for _, m := range m.Methods {
		method := m.Method
		if IsNameParent(scopeName) && method.GetScope() != currentClass && method.VisibilityModifier != Private {
			results = append(results, m)
		} else if IsNameRelative(scopeName) && (method.GetScope() == currentClass || method.VisibilityModifier != Private) {
			results = append(results, m)
		} else if method.IsStatic && method.VisibilityModifier == Public {
			results = append(results, m)
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
func (m InheritedMethods) ReduceAccess(currentClass, scopeName string, types TypeComposite) []MethodWithScope {
	var results []MethodWithScope
	for _, m := range m.Methods {
		method := m.Method
		if scopeName == "$this" && (method.GetScope() == currentClass || method.VisibilityModifier != Private) {
			results = append(results, m)
			continue
		}
		var isSameClass bool
		for _, typeString := range types.Resolve() {
			if typeString.GetFQN() == currentClass {
				isSameClass = true
				break
			}
		}
		if isSameClass && (method.GetScope() == currentClass || method.VisibilityModifier != Private) {
			results = append(results, m)
		} else if method.VisibilityModifier == Public {
			results = append(results, m)
		}
	}
	return results
}

// Len returns number of methods
func (m InheritedMethods) Len() int {
	return len(m.Methods)
}

// GetMethods searches for all methods under the given scope, this function
// does not consider inheritance
func (q *Query) GetMethods(scope string, name string) []*Method {
	cacheKey := "Methods" + sep + scope + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if methods, ok := data.([]*Method); ok {
			return methods
		}
	}
	methods := q.store.GetMethods(scope, name)
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

// GetClassConsts is a cached proxy to store
func (q *Query) GetClassConsts(scope string, name string) []*ClassConst {
	cacheKey := "ClassConst" + sep + scope + "::" + name
	if data, ok := q.cache[cacheKey]; ok {
		if classConsts, ok := data.([]*ClassConst); ok {
			return classConsts
		}
	}
	classConsts := q.store.GetClassConsts(scope, name)
	q.cache[cacheKey] = classConsts
	return classConsts
}

// PropWithScope represents a property with a scope
type PropWithScope struct {
	Prop  *Property
	Scope Symbol
}

func propWithScopeFromProps(scope Symbol, props []*Property) []PropWithScope {
	results := []PropWithScope{}
	for _, prop := range props {
		results = append(results, PropWithScope{prop, scope})
	}
	return results
}

// InheritedProps contains information for props include inheried ones
type InheritedProps struct {
	Props        []PropWithScope
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
		SearchedFQNs: searchedFQNs,
	}
}

// Merge merges the current inherited props with others
func (p *InheritedProps) Merge(other InheritedProps) {
	p.Props = append(p.Props, other.Props...)
	if p.SearchedFQNs == nil {
		p.SearchedFQNs = make(map[string]struct{})
	}
	for fqn := range other.SearchedFQNs {
		p.SearchedFQNs[fqn] = struct{}{}
	}
}

// ReduceStatic reduces properties using the static rules
func (p InheritedProps) ReduceStatic(currentClass, scopeName string) []PropWithScope {
	results := []PropWithScope{}
	for _, ps := range p.Props {
		prop := ps.Prop
		// Properties are different from methods,
		// and static can only be accessed using :: (static::, self::, parent::, TestClass1::)
		if !prop.IsStatic {
			continue
		}
		if IsNameParent(scopeName) && prop.GetScope() != currentClass && prop.VisibilityModifier != Private {
			results = append(results, ps)
		} else if IsNameRelative(scopeName) && (prop.GetScope() == currentClass || prop.VisibilityModifier != Private) {
			results = append(results, ps)
		} else if prop.VisibilityModifier == Public {
			results = append(results, ps)
		}
	}
	return results
}

// ReduceAccess reduces propties using the access rules
func (p InheritedProps) ReduceAccess(currentClass, scopeName string, types TypeComposite) []PropWithScope {
	results := []PropWithScope{}
	for _, ps := range p.Props {
		prop := ps.Prop
		if scopeName == "$this" && (prop.GetScope() == currentClass || prop.VisibilityModifier != Private) {
			results = append(results, ps)
			continue
		}
		isSameClass := false
		for _, typeString := range types.Resolve() {
			if typeString.GetFQN() == currentClass {
				isSameClass = true
				break
			}
		}
		if isSameClass && (prop.GetScope() == currentClass || prop.VisibilityModifier != Private) {
			results = append(results, ps)
		} else if prop.VisibilityModifier == Public {
			results = append(results, ps)
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
	p := q.store.GetProperties(scope, name)
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
