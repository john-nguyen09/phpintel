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
	classMethods := []*Method{}
	if keyword != "" {
		classMethods, _ = store.SearchMethods(class.Name.GetFQN(), keyword, options)
	} else {
		for _, classMethod := range store.GetAllMethods(class.Name.GetFQN()) {
			if isSymbolValid(classMethod, options) {
				classMethods = append(classMethods, classMethod)
			}
		}
	}
	methods = append(methods, classMethods...)

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			if keyword != "" {
				methods = append(methods, searchTraitMethods(store, trait, keyword,
					options)...)
			} else {
				methods = append(methods, getAllTraitMethods(store, trait, options)...)
			}
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			methods = append(methods, SearchClassMethods(store, class, keyword, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			if keyword != "" {
				methods = append(methods, searchInterfaceMethods(store, theInterface, keyword,
					options)...)
			} else {
				methods = append(methods, getAllInterfaceMethods(store, theInterface,
					options)...)
			}
		}
	}
	return methods
}

func GetClassMethods(store *Store, class *Class, name string, options SearchOptions) []*Method {
	methods := []*Method{}
	for _, classMethod := range store.GetMethods(class.Name.GetFQN(), name) {
		if isSymbolValid(classMethod, options) {
			methods = append(methods, classMethod)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			methods = append(methods, getTraitMethods(store, trait, name, options)...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			methods = append(methods, GetClassMethods(store, class, name, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			methods = append(methods, getInterfaceMethods(store, theInterface, name,
				options)...)
		}
	}
	return methods
}

func searchTraitProps(store *Store, trait *Trait, keyword string, options SearchOptions) []*Property {
	props := []*Property{}
	traitProps, _ := store.SearchProperties(trait.Name.GetFQN(), keyword, options)
	props = append(props, traitProps...)
	return props
}

func getAllTraitProps(store *Store, trait *Trait, options SearchOptions) []*Property {
	props := []*Property{}
	for _, traitProp := range store.GetAllProperties(trait.Name.GetFQN()) {
		if isSymbolValid(traitProp, options) {
			props = append(props, traitProp)
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

func searchInterfaceProps(store *Store, theInterface *Interface, keyword string, options SearchOptions) []*Property {
	props := []*Property{}
	interfaceProps, _ := store.SearchProperties(theInterface.Name.GetFQN(), keyword, options)
	props = append(props, interfaceProps...)
	return props
}

func getAllInterfaceProps(store *Store, theInterface *Interface, options SearchOptions) []*Property {
	props := []*Property{}
	for _, interfaceProp := range store.GetAllProperties(theInterface.Name.GetFQN()) {
		if isSymbolValid(interfaceProp, options) {
			props = append(props, interfaceProp)
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

func SearchClassProperties(store *Store, class *Class, keyword string, options SearchOptions) []*Property {
	props := []*Property{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		prop := symbol.(*Property)
		if _, ok := excludeNames[prop.GetName()]; ok {
			return false
		}
		excludeNames[prop.GetName()] = true
		return true
	}
	classProps := []*Property{}
	if keyword != "" {
		classProps, _ = store.SearchProperties(class.Name.GetFQN(), keyword, options.WithPredicate(noDuplicate))
	} else {
		for _, classProp := range store.GetAllProperties(class.Name.GetFQN()) {
			if isSymbolValid(classProp, options.WithPredicate(noDuplicate)) {
				classProps = append(classProps, classProp)
			}
		}
	}
	props = append(props, classProps...)

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			if keyword != "" {
				props = append(props, searchTraitProps(store, trait, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				props = append(props, getAllTraitProps(store, trait, options.WithPredicate(noDuplicate))...)
			}
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			props = append(props, SearchClassProperties(store, class, keyword, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			if keyword != "" {
				props = append(props, searchInterfaceProps(store, theInterface, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				props = append(props, getAllInterfaceProps(store, theInterface,
					options.WithPredicate(noDuplicate))...)
			}
		}
	}
	return props
}

func GetClassProperties(store *Store, class *Class, name string, options SearchOptions) []*Property {
	props := []*Property{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		prop := symbol.(*Property)
		if _, ok := excludeNames[prop.GetName()]; ok {
			return false
		}
		excludeNames[prop.GetName()] = true
		return true
	}
	for _, classProp := range store.GetProperties(class.Name.GetFQN(), name) {
		if isSymbolValid(classProp, options.WithPredicate(noDuplicate)) {
			props = append(props, classProp)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			props = append(props, getTraitProps(store, trait, name, options.WithPredicate(noDuplicate))...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			props = append(props, GetClassProperties(store, class, name, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			props = append(props, getInterfaceProps(store, theInterface, name,
				options.WithPredicate(noDuplicate))...)
		}
	}
	return props
}

func searchTraitClassConsts(store *Store, trait *Trait, keyword string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	traitClassConsts, _ := store.SearchClassConsts(trait.Name.GetFQN(), keyword, options)
	classConsts = append(classConsts, traitClassConsts...)
	return classConsts
}

func getAllTraitClassConsts(store *Store, trait *Trait, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, traitClassConst := range store.GetAllClassConsts(trait.Name.GetFQN()) {
		if isSymbolValid(traitClassConst, options) {
			classConsts = append(classConsts, traitClassConst)
		}
	}
	return classConsts
}

func getTraitClassConsts(store *Store, trait *Trait, name string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, traitClassConst := range store.GetClassConsts(trait.Name.GetFQN(), name) {
		if isSymbolValid(traitClassConst, options) {
			classConsts = append(classConsts, traitClassConst)
		}
	}
	return classConsts
}

func searchInterfaceClassConsts(store *Store, theInterface *Interface, keyword string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	interfaceClassConsts, _ := store.SearchClassConsts(theInterface.Name.GetFQN(), keyword, options)
	classConsts = append(classConsts, interfaceClassConsts...)
	return classConsts
}

func getAllInterfaceClassConsts(store *Store, theInterface *Interface, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, interfaceClassConst := range store.GetAllClassConsts(theInterface.Name.GetFQN()) {
		if isSymbolValid(interfaceClassConst, options) {
			classConsts = append(classConsts, interfaceClassConst)
		}
	}
	return classConsts
}

func getInterfaceClassConsts(store *Store, theInterface *Interface, name string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	for _, interfaceClassConst := range store.GetClassConsts(theInterface.Name.GetFQN(), name) {
		if isSymbolValid(interfaceClassConst, options) {
			classConsts = append(classConsts, interfaceClassConst)
		}
	}
	return classConsts
}

func SearchClassClassConsts(store *Store, class *Class, keyword string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		prop := symbol.(*ClassConst)
		if _, ok := excludeNames[prop.GetName()]; ok {
			return false
		}
		excludeNames[prop.GetName()] = true
		return true
	}
	classClassConsts := []*ClassConst{}
	if keyword != "" {
		classClassConsts, _ = store.SearchClassConsts(class.Name.GetFQN(), keyword, options.WithPredicate(noDuplicate))
	} else {
		for _, classClassConst := range store.GetAllClassConsts(class.Name.GetFQN()) {
			if isSymbolValid(classClassConst, options.WithPredicate(noDuplicate)) {
				classClassConsts = append(classClassConsts, classClassConst)
			}
		}
	}
	classConsts = append(classConsts, classClassConsts...)

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			if keyword != "" {
				classConsts = append(classConsts, searchTraitClassConsts(store, trait, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				classConsts = append(classConsts, getAllTraitClassConsts(store, trait, options.WithPredicate(noDuplicate))...)
			}
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			classConsts = append(classConsts, SearchClassClassConsts(store, class, keyword, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			if keyword != "" {
				classConsts = append(classConsts, searchInterfaceClassConsts(store, theInterface, keyword,
					options.WithPredicate(noDuplicate))...)
			} else {
				classConsts = append(classConsts, getAllInterfaceClassConsts(store, theInterface,
					options.WithPredicate(noDuplicate))...)
			}
		}
	}
	return classConsts
}

func GetClassClassConsts(store *Store, class *Class, name string, options SearchOptions) []*ClassConst {
	classConsts := []*ClassConst{}
	excludeNames := map[string]bool{}
	noDuplicate := func(symbol Symbol) bool {
		prop := symbol.(*ClassConst)
		if _, ok := excludeNames[prop.GetName()]; ok {
			return false
		}
		excludeNames[prop.GetName()] = true
		return true
	}
	for _, classClassConst := range store.GetClassConsts(class.Name.GetFQN(), name) {
		if isSymbolValid(classClassConst, options.WithPredicate(noDuplicate)) {
			classConsts = append(classConsts, classClassConst)
		}
	}

	for _, traitName := range class.Use {
		if traitName.IsEmpty() {
			continue
		}
		for _, trait := range store.GetTraits(traitName.GetFQN()) {
			classConsts = append(classConsts, getTraitClassConsts(store, trait, name, options.WithPredicate(noDuplicate))...)
		}
	}

	if !class.Extends.IsEmpty() {
		for _, class := range store.GetClasses(class.Extends.GetFQN()) {
			classConsts = append(classConsts, GetClassClassConsts(store, class, name, options)...)
		}
	}

	for _, typeString := range class.Interfaces {
		if typeString.IsEmpty() {
			continue
		}
		for _, theInterface := range store.GetInterfaces(typeString.GetFQN()) {
			classConsts = append(classConsts, getInterfaceClassConsts(store, theInterface, name,
				options.WithPredicate(noDuplicate))...)
		}
	}
	return classConsts
}
