package analysis

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestGlobalVariable(t *testing.T) {
	globalVariableTest := "../cases/globalVariable.php"
	data, _ := ioutil.ReadFile(globalVariableTest)
	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestGlobalVariableReference(t *testing.T) {
	store, _ := NewStore("./testData/globalVariable")
	globalVariableTest, _ := filepath.Abs("../cases/globalVariable.php")
	indexDocument(store, globalVariableTest, "test1")

	referenceFile, _ := filepath.Abs("../cases/reference/globalVariable.php")
	document := openDocument(store, referenceFile, "test2")

	symbol := document.SymbolAt(14)
	var propAccess *PropertyAccess = nil
	var ok bool
	if propAccess, ok = symbol.(*PropertyAccess); !ok {
		t.Errorf("At 14, %T is not *PropertyAccess", symbol)
	}
	testResult := struct {
		dbTypes     TypeComposite
		outputTypes TypeComposite
		varTypes    TypeComposite

		insideFunctionNoGlobalTypes  TypeComposite
		insideFunctionHasGlobalTypes TypeComposite
	}{
		dbTypes:     newTypeComposite(),
		outputTypes: newTypeComposite(),
		varTypes:    newTypeComposite(),

		insideFunctionNoGlobalTypes:  newTypeComposite(),
		insideFunctionHasGlobalTypes: newTypeComposite(),
	}
	testResult.dbTypes = propAccess.ResolveAndGetScope(store)

	symbol = document.SymbolAt(28)
	if propAccess, ok = symbol.(*PropertyAccess); !ok {
		t.Errorf("At 28, %T is not *PropertyAccess", symbol)
	}
	testResult.outputTypes = propAccess.ResolveAndGetScope(store)

	symbol = document.SymbolAt(35)
	var variable *Variable = nil
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 35, %T is not *Variable", symbol)
	}
	variable.Resolve(store)
	testResult.varTypes = variable.GetTypes()

	symbol = document.SymbolAt(109)
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 109, %T is not *Variable", symbol)
	}
	variable.Resolve(store)
	testResult.insideFunctionNoGlobalTypes = variable.GetTypes()

	symbol = document.SymbolAt(177)
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 177, %T is not *Variable", symbol)
	}
	variable.Resolve(store)
	testResult.insideFunctionHasGlobalTypes = variable.GetTypes()

	cupaloy.SnapshotT(t, testResult)
}
