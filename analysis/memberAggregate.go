package analysis

func (s *Class) SearchInheritedMethods(store *Store, keyword string, excludedMethods []*Method) []*Method {
	methods := []*Method{}
	excludeNames := map[string]bool{}
	for _, excludedMethod := range excludedMethods {
		excludeNames[excludedMethod.Name] = true
	}
	for _, method := range store.SearchMethods(s.Extends.GetFQN(), keyword) {
		if _, ok := excludeNames[method.Name]; ok || method.VisibilityModifier == Private {
			continue
		}
		methods = append(methods, method)
		excludeNames[method.Name] = true
	}
	for _, typeString := range s.Interfaces {
		for _, method := range store.SearchMethods(typeString.GetFQN(), keyword) {
			if _, ok := excludeNames[method.Name]; ok {
				continue
			}
			methods = append(methods, method)
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
	for _, method := range store.GetMethods(s.Extends.GetFQN(), name) {
		if _, ok := excludeNames[method.Name]; ok || method.VisibilityModifier == Private {
			continue
		}
		methods = append(methods, method)
		excludeNames[method.Name] = true
	}
	for _, typeString := range s.Interfaces {
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
	for _, property := range store.SearchProperties(s.Extends.GetFQN(), keyword) {
		if _, ok := excludeNames[property.Name]; ok {
			continue
		}
		properties = append(properties, property)
		excludeNames[property.Name] = true
	}
	for _, typeString := range s.Interfaces {
		for _, property := range store.SearchProperties(typeString.GetFQN(), keyword) {
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
	for _, property := range store.GetProperties(s.Extends.GetFQN(), name) {
		if _, ok := excludeNames[property.Name]; ok {
			continue
		}
		properties = append(properties, property)
		excludeNames[property.Name] = true
	}
	for _, typeString := range s.Interfaces {
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
	for _, classConst := range store.SearchClassConsts(s.Extends.GetFQN(), keyword) {
		if _, ok := excludeNames[classConst.Name]; ok {
			continue
		}
		classConsts = append(classConsts, classConst)
		excludeNames[classConst.Name] = true
	}
	for _, typeString := range s.Interfaces {
		for _, classConst := range store.SearchClassConsts(typeString.GetFQN(), keyword) {
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
	for _, classConst := range store.GetClassConsts(s.Extends.GetFQN(), name) {
		if _, ok := excludeNames[classConst.Name]; ok {
			continue
		}
		classConsts = append(classConsts, classConst)
		excludeNames[classConst.Name] = true
	}
	for _, typeString := range s.Interfaces {
		for _, classConst := range store.GetClassConsts(typeString.GetFQN(), name) {
			if _, ok := excludeNames[classConst.Name]; ok {
				continue
			}
			classConsts = append(classConsts, classConst)
		}
	}
	return classConsts
}
