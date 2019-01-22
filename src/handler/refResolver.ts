import { Reference, RefKind } from "../symbol/reference";
import { App } from "../app";
import { FunctionTable } from "../storage/table/function";
import { TypeComposite } from "../type/composite";
import { Function } from "../symbol/function/function";
import { PhpDocument } from "../symbol/phpDocument";
import { Class } from "../symbol/class/class";
import { ClassTable } from "../storage/table/class";
import { Method } from "../symbol/function/method";
import { MethodTable } from "../storage/table/method";
import { Property } from "../symbol/variable/property";
import { PropertyTable } from "../storage/table/property";
import { ClassConstant } from "../symbol/constant/classConstant";
import { ClassConstantTable } from "../storage/table/classConstant";
import { Constant } from "../symbol/constant/constant";
import { ConstantTable } from "../storage/table/constant";
import { Symbol } from "../symbol/symbol";
import { CompletionValue } from "../storage/table/index/completionIndex";
import { SymbolModifier } from "../symbol/meta/modifier";
import { ScopeVarTable } from "../storage/table/scopeVar";
import { ReferenceTable } from "../storage/table/reference";
import { Variable } from "../symbol/variable/variable";
import { TypeName } from "../type/name";

export namespace RefResolver {
    export async function getSymbolsByReference(phpDoc: PhpDocument, ref: Reference): Promise<Symbol[]> {
        let symbols: Symbol[] = [];

        switch (ref.refKind) {
            case RefKind.Function:
                symbols = await RefResolver.getFuncSymbols(phpDoc, ref);
                break;
            case RefKind.ClassTypeDesignator:
                symbols = await RefResolver.getMethodSymbols(phpDoc, ref);

                if (symbols.length === 0) {
                    symbols = await RefResolver.getClassSymbols(phpDoc, ref);
                }
                break;
            case RefKind.Class:
                symbols = await RefResolver.getClassSymbols(phpDoc, ref);
                break;
            case RefKind.Method:
                symbols = await RefResolver.getMethodSymbols(phpDoc, ref);
                break;
            case RefKind.Property:
                symbols = await RefResolver.getPropSymbols(phpDoc, ref);
                break;
            case RefKind.ClassConst:
                symbols = await RefResolver.getClassConstSymbols(phpDoc, ref);
                break;
            case RefKind.ConstantAccess:
                symbols = await RefResolver.getConstSymbols(phpDoc, ref);
                break;
        }

        return symbols;
    }

    export async function searchSymbolsForReference(phpDoc: PhpDocument, ref: Reference): Promise<Symbol[]> {
        const funcTable = App.get<FunctionTable>(FunctionTable);
        const classTable = App.get<ClassTable>(ClassTable);
        const constTable = App.get<ConstantTable>(ConstantTable);
        const methodTable = App.get<MethodTable>(MethodTable);
        const propTable = App.get<PropertyTable>(PropertyTable);
        const classConstTable = App.get<ClassConstantTable>(ClassConstantTable);
        const scopeVarTable = App.get<ScopeVarTable>(ScopeVarTable);
        const refTable = App.get<ReferenceTable>(ReferenceTable);

        let symbols: Symbol[] = [];
        let keyword: string;
        let scopeName: string;
        let completions: CompletionValue[];

        switch (ref.refKind) {
            case RefKind.ConstantAccess:
                keyword = ref.type.toString();
                completions = await funcTable.search(keyword);
                for (let completion of completions) {
                    symbols.push(...await funcTable.get(completion.name));
                }

                completions = await classTable.search(keyword);
                for (let completion of completions) {
                    symbols.push(...await classTable.get(completion.name));
                }

                completions = await constTable.search(keyword);
                for (let completion of completions) {
                    symbols.push(...await constTable.get(completion.name));
                }

                break;
            case RefKind.ClassConst:
            case RefKind.ScopedAccess:
                keyword = ref.type.toString();
                if (ref.scope === null) {
                    break;
                }
                ref.scope.resolveReferenceToFqn(phpDoc.importTable);
                scopeName = ref.scope.toString();

                if (keyword.length > 0) {
                    completions = await methodTable.search(scopeName, keyword);
                    for (let completion of completions) {
                        symbols.push(...await methodTable.getByClass(scopeName, completion.name));
                    }

                    completions = await classConstTable.search(scopeName, keyword);
                    for (let completion of completions) {
                        symbols.push(...await classConstTable.getByClass(scopeName, completion.name));
                    }
                } else {
                    symbols.push(...await methodTable.searchAllInClass(scopeName, (method) => {
                        return method.modifier.has(SymbolModifier.STATIC);
                    }));
                    symbols.push(...await propTable.searchAllInClass(scopeName, (prop) => {
                        return prop.modifier.has(SymbolModifier.STATIC);
                    }));
                    symbols.push(...await classConstTable.searchAllInClass(scopeName));
                }

                break;
            case RefKind.Property:
                keyword = ref.type.toString();
                if (ref.scope === null) {
                    break;
                }
                ref.scope.resolveReferenceToFqn(phpDoc.importTable);
                scopeName = ref.scope.toString();

                const isRefStatic = typeof ref.refName === 'undefined';
                if (keyword.length > 0) {
                    completions = await propTable.search(scopeName, keyword);
                    for (let completion of completions) {
                        symbols.push(...(await propTable.getByClass(scopeName, completion.name))
                            .filter((prop) => {
                                return !isRefStatic || prop.modifier.has(SymbolModifier.STATIC);
                            }));
                    }
                } else {
                    symbols.push(...await propTable.searchAllInClass(scopeName, (prop) => {
                        return !isRefStatic || prop.modifier.has(SymbolModifier.STATIC);
                    }));
                }

                break;
            case RefKind.Variable:
                keyword = '';
                if (typeof ref.refName !== 'undefined') {
                    keyword = ref.refName;
                }
                const range = await scopeVarTable.findAt(ref.location.uri, ref.location.range.start);
                console.log(range);

                if (range === null) {
                    break;
                }

                let refVars: Reference[] = [];
                if (keyword.length > 0) {

                } else {
                    refVars = await refTable.findWithin(phpDoc.uri, range, (foundRef) => {
                        return foundRef.refKind === RefKind.Variable &&
                            typeof foundRef.refName !== 'undefined' &&
                            foundRef.refName.length > 0;
                    });
                }
                if (refVars.length > 0) {
                    for (let refVar of refVars) {
                        if (typeof refVar.refName == 'undefined' || refVar.type instanceof TypeName) {
                            continue;
                        }

                        symbols.push(new Variable(refVar.refName, refVar.type));
                    }
                }

                break;
        }

        return symbols;
    }

    export async function getFuncSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Function[]> {
        const funcTable = App.get<FunctionTable>(FunctionTable);

        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveReferenceToFqn(phpDoc.importTable);

        return await funcTable.get(ref.type.name);
    }

    export async function getMethodSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Method[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const methodTable = App.get<MethodTable>(MethodTable);
        let className: string = '';
        let methodName: string = '';

        if (ref.refKind === RefKind.ClassTypeDesignator) {
            ref.type.resolveReferenceToFqn(phpDoc.importTable);
            className = ref.type.name;
            methodName = '__construct';
        } else if (ref.refKind === RefKind.Method && ref.scope !== null) {
            ref.scope.resolveReferenceToFqn(phpDoc.importTable);
            className = ref.scope.name;
            methodName = ref.type.name;
        }

        return await methodTable.getByClass(className, methodName);
    }

    export async function getPropSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Property[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const propTable = App.get<PropertyTable>(PropertyTable);
        let className = '';

        if (ref.scope !== null) {
            ref.scope.resolveReferenceToFqn(phpDoc.importTable);
            className = ref.scope.name;
        }

        return await propTable.getByClass(className, ref.type.name);
    }

    export async function getClassSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Class[]> {
        const classTable = App.get<ClassTable>(ClassTable);

        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveReferenceToFqn(phpDoc.importTable);

        return await classTable.get(ref.type.name);
    }

    export async function getClassConstSymbols(phpDoc: PhpDocument, ref: Reference): Promise<ClassConstant[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const classConstTable = App.get<ClassConstantTable>(ClassConstantTable);
        let className = '';

        if (ref.scope !== null) {
            ref.scope.resolveReferenceToFqn(phpDoc.importTable);
            className = ref.scope.name;
        }

        return await classConstTable.getByClass(className, ref.type.name);
    }

    export async function getConstSymbols(
        phpDoc: PhpDocument,
        ref: Reference
    ): Promise<Constant[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        ref.type.resolveReferenceToFqn(phpDoc.importTable);

        const constTable = App.get<ConstantTable>(ConstantTable);

        return constTable.get(ref.type.name);
    }
}