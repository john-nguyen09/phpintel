package analysis

func SearchTraitMethods(store *Store, trait *Trait, keyword string, options SearchOptions) []*Method {
	methods := []*Method{}
	traitMethods, _ := store.SearchMethods(trait.Name.GetFQN(), keyword, options)
	methods = append(methods, traitMethods...)
	return methods
}

func SearchInterfaceMethods(store *Store, theInterface *Interface, keyword string,
	options SearchOptions) []*Method {
	methods := []*Method{}
	interfaceMethods, _ := store.SearchMethods(theInterface.Name.GetFQN(), keyword, options)
	methods = append(methods, interfaceMethods...)
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
	classMethods, _ := store.SearchMethods(class.Name.GetFQN(), keyword, options.WithPredicate(noDuplicate))
	methods = append(methods, classMethods...)

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		traits := store.GetTraits(traitName.GetFQN())
		for _, trait := range traits {
			methods = append(methods, SearchTraitMethods(store, trait, keyword,
				options.WithPredicate(noDuplicate))...)
		}
	}

	if !class.Extends.IsEmpty() {
		classes := store.GetClasses(class.Name.GetFQN())
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
			methods = append(methods, SearchInterfaceMethods(store, theInterface, keyword,
				options.WithPredicate(noDuplicate))...)
		}
	}
	return methods
}

func (s *Class) GetInheritedMethods(store *Store, name string, excludedMethods []*Method) []*Method {
	methods := []*Method{}
	excludeNames := map[string]bool{}
	for _, excludedMethod := range excludedMethods {
		excludeNames[excludedMethod.Name] = true
	}
	for _, typeString := range s.Use {
		if typeString.IsEmpty() {
			continue
		}
		for _, method := range store.GetMethods(typeString.GetFQN(), name) {
			if _, ok := excludeNames[method.Name]; ok {
				continue
			}
			methods = append(methods, method)
			excludeNames[method.Name] = true
		}
	}
	if !s.Extends.IsEmpty() {
		for _, method := range store.GetMethods(s.Extends.GetFQN(), name) {
			if _, ok := excludeNames[method.Name]; ok || method.VisibilityModifier == Private {
				continue
			}
			methods = append(methods, method)
			excludeNames[method.Name] = true
		}
	}
	for _, typeString := range s.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, method := range store.GetMethods(typeString.GetFQN(), name) {
			if _, ok := excludeNames[method.Name]; ok {
				continue
			}
			methods = append(methods, method)
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
