package analysis

import (
	"strings"
)

func withNoDuplicateNamesOptions(opt SearchOptions) SearchOptions {
	excludeNames := map[string]bool{}
	return opt.
		WithPredicate(func(symbol Symbol) bool {
			hasName := symbol.(HasName)
			key := hasName.GetName()
			if _, ok := excludeNames[key]; ok {
				return false
			}
			excludeNames[key] = true
			return true
		})
}

func withKeywordOptions(keyword string, opt SearchOptions) SearchOptions {
	if keyword == "" {
		return opt
	}
	keyword = strings.ToLower(keyword)
	return opt.
		WithPredicate(func(symbol Symbol) bool {
			if v, ok := symbol.(NameIndexable); ok {
				return strings.Contains(strings.ToLower(v.GetIndexableName()), keyword)
			}
			return true
		})
}

func SearchTraitMethods(store *Store, trait *Trait, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, method := range store.GetAllMethods(trait.Name.GetFQN()) {
		if isSymbolValid(method, options) {
			methods = append(methods, method)
		}
	}
	return methods
}

func GetTraitMethods(store *Store, trait *Trait, name string, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, traitMethod := range store.GetMethods(trait.Name.GetFQN(), name) {
		if isSymbolValid(traitMethod, options) {
			methods = append(methods, traitMethod)
		}
	}
	return methods
}

func searchInterfaceMethods(store *Store, theInterface *Interface, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, method := range store.GetAllMethods(theInterface.Name.GetFQN()) {
		if isSymbolValid(method, options) {
			methods = append(methods, method)
		}
	}

	for _, typeString := range theInterface.Extends {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			methods = append(methods, searchInterfaceMethods(store, theInterface, options)...)
		}
	}

	return methods
}

func SearchInterfaceMethods(store *Store, theInterface *Interface, keyword string, options SearchOptions) []*Method {
	options = withKeywordOptions(keyword, options)
	return searchInterfaceMethods(store, theInterface, options)
}

func searchClassMethods(store *Store, class *Class, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, method := range store.GetAllMethods(class.Name.GetFQN()) {
		if isSymbolValid(method, options) {
			methods = append(methods, method)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			methods = append(methods, SearchTraitMethods(store, trait, options)...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			methods = append(methods, searchClassMethods(store, class, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			methods = append(methods, searchInterfaceMethods(store, theInterface, options)...)
		}
	}
	return methods
}

func SearchClassMethods(store *Store, class *Class, keyword string, options SearchOptions) []*Method {
	options = withKeywordOptions(keyword, withNoDuplicateNamesOptions(options))
	return searchClassMethods(store, class, options)
}

func searchTraitProps(store *Store, trait *Trait, options SearchOptions) []*Property {
	props := []*Property{}
	for _, prop := range store.GetAllProperties(trait.Name.GetFQN()) {
		if isSymbolValid(prop, options) {
			props = append(props, prop)
		}
	}
	return props
}

func getTraitProps(store *Store, trait *Trait, name string, options SearchOptions) []*Property {
	props := []*Property{}
	for _, traitProp := range store.GetProperties(trait.Name.GetFQN(), name) {
		if isSymbolValid(traitProp, options) {
			props = append(props, traitProp)
		}
	}
	return props
}

func searchInterfaceProps(store *Store, theInterface *Interface, options SearchOptions) []*Property {
	props := []*Property{}
	for _, prop := range store.GetAllProperties(theInterface.Name.GetFQN()) {
		if isSymbolValid(prop, options) {
			props = append(props, prop)
		}
	}
	return props
}

func getInterfaceProps(store *Store, theInterface *Interface, name string, options SearchOptions) []*Property {
	props := []*Property{}
	for _, interfaceProp := range store.GetProperties(theInterface.Name.GetFQN(), name) {
		if isSymbolValid(interfaceProp, options) {
			props = append(props, interfaceProp)
		}
	}
	return props
}

func searchClassProperties(store *Store, class *Class, options SearchOptions) []*Property {
	props := []*Property{}
	for _, prop := range store.GetAllProperties(class.Name.GetFQN()) {
		if isSymbolValid(prop, options) {
			props = append(props, prop)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			props = append(props, searchTraitProps(store, trait, options)...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			props = append(props, searchClassProperties(store, class, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			props = append(props, searchInterfaceProps(store, theInterface, options)...)
		}
	}
	return props
}

func SearchClassProperties(store *Store, class *Class, keyword string, options SearchOptions) []*Property {
	options = withKeywordOptions(keyword, withNoDuplicateNamesOptions(options))
	return searchClassProperties(store, class, options)
}

func searchTraitClassConsts(store *Store, trait *Trait, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, classConst := range store.GetAllClassConsts(trait.Name.GetFQN()) {
		if isSymbolValid(classConst, options) {
			classConsts = append(classConsts, classConst)
		}
	}
	return classConsts
}

func searchInterfaceClassConsts(store *Store, theInterface *Interface, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, classConst := range store.GetAllClassConsts(theInterface.Name.GetFQN()) {
		if isSymbolValid(classConst, options) {
			classConsts = append(classConsts, classConst)
		}
	}
	return classConsts
}

func searchClassClassConsts(store *Store, class *Class, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, classConst := range store.GetAllClassConsts(class.Name.GetFQN()) {
		if isSymbolValid(classConst, options) {
			classConsts = append(classConsts, classConst)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			classConsts = append(classConsts, searchTraitClassConsts(store, trait, options)...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			classConsts = append(classConsts, searchClassClassConsts(store, class, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			classConsts = append(classConsts, searchInterfaceClassConsts(store, theInterface, options)...)
		}
	}
	return classConsts
}

func SearchClassClassConsts(store *Store, class *Class, keyword string, options SearchOptions) []*ClassConst {
	options = withKeywordOptions(keyword, withNoDuplicateNamesOptions(options))
	return searchClassClassConsts(store, class, options)
}

func StaticMethodsScopeAware(opt SearchOptions, classScope string, name string) SearchOptions {
	return opt.WithPredicate(func(symbol Symbol) bool {
		method := symbol.(*Method)
		if IsNameParent(name) {
			// parent:: excludes methods from current class
			if method.GetScope() == classScope {
				return false
			}
			// or from parents but private
			if method.VisibilityModifier == Private {
				return false
			}
			return true
		}
		// static:: and self:: exclude private methods that are not from current class
		if IsNameRelative(name) {
			if method.GetScope() != classScope &&
				method.VisibilityModifier == Private {
				return false
			}
			// And also accept non-static
			return true
		}
		// Not parent:: or static:: or self:: so accept only public static
		return method.IsStatic && method.VisibilityModifier == Public
	})
}

func MethodsScopeAware(opt SearchOptions, document *Document, scope HasTypes) SearchOptions {
	name := ""
	classScope := document.GetClassScopeAtSymbol(scope)
	if hasName, ok := scope.(HasName); ok {
		name = hasName.GetName()
	}
	return opt.WithPredicate(func(symbol Symbol) bool {
		method := symbol.(*Method)
		// $this allows excludes private methods from parents
		if name == "$this" {
			if method.GetScope() != classScope && method.VisibilityModifier == Private {
				return false
			}
			return true
		}
		// The same goes for the type of the same class not just $this
		isSameClass := false
		for _, typeString := range scope.GetTypes().Resolve() {
			if typeString.GetFQN() == classScope {
				isSameClass = true
				break
			}
		}
		if isSameClass {
			if method.GetScope() != classScope && method.VisibilityModifier == Private {
				return false
			}
			return true
		}
		return method.VisibilityModifier == Public
	})
}

func StaticPropsScopeAware(opt SearchOptions, classScope string, name string) SearchOptions {
	return opt.WithPredicate(func(symbol Symbol) bool {
		prop := symbol.(*Property)
		// Properties are different from methods,
		// and static can only be accessed using :: (static::, self::, parent::, TestClass1::)
		if !prop.IsStatic {
			return false
		}
		if IsNameParent(name) {
			if prop.GetScope() == classScope || prop.VisibilityModifier == Private {
				return false
			}
			return true
		}
		if IsNameRelative(name) {
			if prop.GetScope() != classScope && prop.VisibilityModifier == Private {
				return false
			}
			return true
		}
		return prop.VisibilityModifier == Public
	})
}

func PropsScopeAware(opt SearchOptions, document *Document, scope HasTypes) SearchOptions {
	name := ""
	classScope := document.GetClassScopeAtSymbol(scope)
	if hasName, ok := scope.(HasName); ok {
		name = hasName.GetName()
	}
	return opt.WithPredicate(func(symbol Symbol) bool {
		prop := symbol.(*Property)
		if name == "$this" {
			if prop.GetScope() != classScope && prop.VisibilityModifier == Private {
				return false
			}
			return true
		}
		isSameClass := false
		for _, typeString := range scope.GetTypes().Resolve() {
			if typeString.GetFQN() == classScope {
				isSameClass = true
				break
			}
		}
		if isSameClass {
			if prop.GetScope() != classScope && prop.VisibilityModifier == Private {
				return false
			}
			return true
		}
		return prop.VisibilityModifier == Public
	})
}
