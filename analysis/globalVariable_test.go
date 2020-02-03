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
	store, _ := setupStore("test", "globalVariable")
	globalVariableTest, _ := filepath.Abs("../cases/globalVariable.php")
	indexDocument(store, globalVariableTest, "test1")

	referenceFile, _ := filepath.Abs("../cases/reference/globalVariable.php")
	document := openDocument(store, referenceFile, "test2")
	resolveCtx := NewResolveContext(store, document)

	symbol := document.HasTypesAt(14)
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
	testResult.dbTypes = propAccess.ResolveAndGetScope(resolveCtx)

	symbol = document.HasTypesAt(28)
	if propAccess, ok = symbol.(*PropertyAccess); !ok {
		t.Errorf("At 28, %T is not *PropertyAccess", symbol)
	}
	testResult.outputTypes = propAccess.ResolveAndGetScope(resolveCtx)

	symbol = document.HasTypesAt(35)
	var variable *Variable = nil
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 35, %T is not *Variable", symbol)
	}
	variable.Resolve(resolveCtx)
	testResult.varTypes = variable.GetTypes()

	symbol = document.HasTypesAt(109)
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 109, %T is not *Variable", symbol)
	}
	variable.Resolve(resolveCtx)
	testResult.insideFunctionNoGlobalTypes = variable.GetTypes()

	symbol = document.HasTypesAt(177)
	if variable, ok = symbol.(*Variable); !ok {
		t.Errorf("At 177, %T is not *Variable", symbol)
	}
	variable.Resolve(resolveCtx)
	testResult.insideFunctionHasGlobalTypes = variable.GetTypes()

	cupaloy.SnapshotT(t, testResult)
}
