package analysis

func searchTraitMethods(store *Store, trait *Trait, keyword string, options SearchOptions) []*Method {
	methods := []*Method{}
	traitMethods, _ := store.SearchMethods(trait.Name.GetFQN(), keyword, options)
	methods = append(methods, traitMethods...)
	return methods
}

func getAllTraitMethods(store *Store, trait *Trait, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, traitMethod := range store.GetAllMethods(trait.Name.GetFQN()) {
		if isSymbolValid(traitMethod, options) {
			methods = append(methods, traitMethod)
		}
	}
	return methods
}

func getTraitMethods(store *Store, trait *Trait, name string, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, traitMethod := range store.GetMethods(trait.Name.GetFQN(), name) {
		if isSymbolValid(traitMethod, options) {
			methods = append(methods, traitMethod)
		}
	}
	return methods
}

func searchInterfaceMethods(store *Store, theInterface *Interface, keyword string,
	options SearchOptions) []*Method {
	methods := []*Method{}
	interfaceMethods, _ := store.SearchMethods(theInterface.Name.GetFQN(), keyword, options)
	methods = append(methods, interfaceMethods...)
	return methods
}

func getAllInterfaceMethods(store *Store, theInterface *Interface, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, interfaceMethod := range store.GetAllMethods(theInterface.Name.GetFQN()) {
		if isSymbolValid(interfaceMethod, options) {
			methods = append(methods, interfaceMethod)
		}
	}
	return methods
}

func getInterfaceMethods(store *Store, theInterface *Interface, name string, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, interfaceMethod := range store.GetMethods(theInterface.Name.GetFQN(), name) {
		if isSymbolValid(interfaceMethod, options) {
			methods = append(methods, interfaceMethod)
		}
	}
	return methods
}

func SearchClassMethods(store *Store, class *Class, keyword string, options SearchOptions) []*Method {
	methods := []*Method{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		method := symbol.(*Method)
		if _, ok := excludeNames[method.GetName()]; ok {
			return false
		}
		excludeNames[method.GetName()] = true
		return true
	}
	classMethods := []*Method{}
	if keyword != "" {
		classMethods, _ = store.SearchMethods(class.Name.GetFQN(), keyword, options.WithPredicate(noDuplicate))
	} else {
		for _, classMethod := range store.GetAllMethods(class.Name.GetFQN()) {
			if isSymbolValid(classMethod, options.WithPredicate(noDuplicate)) {
				classMethods = append(classMethods, classMethod)
			}
		}
	}
	methods = append(methods, classMethods...)

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		traits := store.GetTraits(traitName.GetFQN())
		for _, trait := range traits {
			if keyword != "" {
				methods = append(methods, searchTraitMethods(store, trait, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				methods = append(methods, getAllTraitMethods(store, trait, options.WithPredicate(noDuplicate))...)
			}
		}
	}

	if !class.Extends.IsEmpty() {
		classes := store.GetClasses(class.Extends.GetFQN())
		for _, class := range classes {
			methods = append(methods, SearchClassMethods(store, class, keyword, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		interfaces := store.GetInterfaces(typeString.GetFQN())
		for _, theInterface := range interfaces {
			if keyword != "" {
				methods = append(methods, searchInterfaceMethods(store, theInterface, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				methods = append(methods, getAllInterfaceMethods(store, theInterface,
					options.WithPredicate(noDuplicate))...)
			}
		}
	}
	return methods
}

func GetClassMethods(store *Store, class *Class, name string, options SearchOptions) []*Method {
	methods := []*Method{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		method := symbol.(*Method)
		if _, ok := excludeNames[method.GetName()]; ok {
			return false
		}
		excludeNames[method.GetName()] = true
		return true
	}
	for _, classMethod := range store.GetMethods(class.Name.GetFQN(), name) {
		if isSymbolValid(classMethod, options.WithPredicate(noDuplicate)) {
			methods = append(methods, classMethod)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		traits := store.GetTraits(traitName.GetFQN())
		for _, trait := range traits {
			methods = append(methods, getTraitMethods(store, trait, name, options.WithPredicate(noDuplicate))...)
		}
	}

	if !class.Extends.IsEmpty() {
		classes := store.GetClasses(class.Extends.GetFQN())
		for _, class := range classes {
			methods = append(methods, GetClassMethods(store, class, name, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		interfaces := store.GetInterfaces(typeString.GetFQN())
		for _, theInterface := range interfaces {
			methods = append(methods, getInterfaceMethods(store, theInterface, name,
				options.WithPredicate(noDuplicate))...)
		}
	}
	return methods
}

func (s *Class) SearchInheritedProperties(store *Store, keyword string, excludedProperties []*Property) []*Property {
	properties := []*Property{}
	excludeNames := map[string]bool{}
	for _, excludedProperty := range excludedProperties {
		excludeNames[excludedProperty.Name] = true
	}
	if !s.Extends.IsEmpty() {
		properties, _ = store.SearchProperties(s.Extends.GetFQN(), keyword, NewSearchOptions())
		for _, property := range properties {
			if _, ok := excludeNames[property.Name]; ok {
				continue
			}
			properties = append(properties, property)
			excludeNames[property.Name] = true
		}
	}
	for _, typeString := range s.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		properties, _ = store.SearchProperties(typeString.GetFQN(), keyword, NewSearchOptions())
		for _, property := range properties {
			if _, ok := excludeNames[property.Name]; ok {
				continue
			}
			properties = append(properties, property)
		}
	}
	return properties
}

func (s *Class) GetInheritedProperties(store *Store, name string, excludedProperties []*Property) []*Property {
	properties := []*Property{}
	excludeNames := map[string]bool{}
	for _, excludedProperty := range excludedProperties {
		excludeNames[excludedProperty.Name] = true
	}
	if !s.Extends.IsEmpty() {
		for _, property := range store.GetProperties(s.Extends.GetFQN(), name) {
			if _, ok := excludeNames[property.Name]; ok {
				continue
			}
			properties = append(properties, property)
			excludeNames[property.Name] = true
		}
	}
	for _, typeString := range s.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, property := range store.GetProperties(typeString.GetFQN(), name) {
			if _, ok := excludeNames[property.Name]; ok {
				continue
			}
			properties = append(properties, property)
		}
	}
	return properties
}

func (s *Class) SearchInheritedClassConsts(store *Store, keyword string) []*ClassConst {
	classConsts := []*ClassConst{}
	excludeNames := map[string]bool{}
	if !s.Extends.IsEmpty() {
		classConsts, _ = store.SearchClassConsts(s.Extends.GetFQN(), keyword, NewSearchOptions())
		for _, classConst := range classConsts {
			if _, ok := excludeNames[classConst.Name]; ok {
				continue
			}
			classConsts = append(classConsts, classConst)
			excludeNames[classConst.Name] = true
		}
	}
	for _, typeString := range s.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		classConsts, _ = store.SearchClassConsts(typeString.GetFQN(), keyword, NewSearchOptions())
		for _, classConst := range classConsts {
			if _, ok := excludeNames[classConst.Name]; ok {
				continue
			}
			classConsts = append(classConsts, classConst)
		}
	}
	return classConsts
}

func (s *Class) GetInheritedClassConsts(store *Store, name string) []*ClassConst {
	classConsts := []*ClassConst{}
	excludeNames := map[string]bool{}
	if !s.Extends.IsEmpty() {
		for _, classConst := range store.GetClassConsts(s.Extends.GetFQN(), name) {
			if _, ok := excludeNames[classConst.Name]; ok {
				continue
			}
			classConsts = append(classConsts, classConst)
			excludeNames[classConst.Name] = true
		}
	}
	for _, typeString := range s.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, classConst := range store.GetClassConsts(typeString.GetFQN(), name) {
			if _, ok := excludeNames[classConst.Name]; ok {
				continue
			}
			classConsts = append(classConsts, classConst)
		}
	}
	return classConsts
}
