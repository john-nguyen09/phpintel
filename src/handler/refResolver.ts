import { Reference, RefKind } from "../symbol/reference";
import { App } from "../app";
import { FunctionTable } from "../storage/table/function";
import { TypeComposite, ResolveType } from "../type/composite";
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
import { ScopeVar } from "../symbol/variable/scopeVar";
import { SignatureHelp, SignatureInformation } from "vscode-languageserver";
import { Formatter } from "./formatter";

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
            case RefKind.MethodCall:
                symbols = await RefResolver.getMethodSymbols(phpDoc, ref);
                break;
            case RefKind.Property:
            case RefKind.PropertyAccess:
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

    export async function searchSymbolsForReference(
        phpDoc: PhpDocument, ref: Reference, offset: number
    ): Promise<Symbol[]> {
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
                if (ref.scope instanceof TypeName) {
                    ref.scope.resolveReferenceToFqn(phpDoc.importTable);
                }
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
                if (ref.scope instanceof TypeName) {
                    ref.scope.resolveReferenceToFqn(phpDoc.importTable);
                }
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
                if (ref.location.uri === undefined || ref.location.range === undefined) {
                    break;
                }
                const range = await scopeVarTable.findAt(ref.location.uri, ref.location.range.start);

                if (range === null) {
                    break;
                }

                let refVars: Reference[] = [];
                const scopeVar = new ScopeVar();
                if (keyword.length > 0) {
                    refVars = await refTable.findWithin(phpDoc.uri, range, (foundRef) => {
                        return foundRef.refKind === RefKind.Variable &&
                            typeof foundRef.refName !== 'undefined' &&
                            foundRef.refName.length > 0 &&
                            foundRef.refName.indexOf(keyword) >= 0;
                    });
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

                        scopeVar.set(new Variable(refVar.refName, refVar.type));
                    }

                    for (let varName in scopeVar.variables) {
                        symbols.push(new Variable(varName, scopeVar.variables[varName]));
                    }
                }

                break;
            case RefKind.PropertyAccess:
                if (ref.scope === null) {
                    break;
                }

                keyword = getReferenceKeyword(ref, offset);
                const scopeNames: string[] = [];

                ResolveType.forType(ref.scope, (scope) => {
                    scope.resolveReferenceToFqn(phpDoc.importTable);
                    scopeNames.push(scope.name);
                });

                if (keyword.length > 0) {
                    const promises: Promise<Symbol[]>[] = [];
                    for (const scopeName of scopeNames) {
                        if (!keyword.startsWith('$')) {
                            completions = await propTable.search(scopeName, '$' + keyword);
                            for (const completion of completions) {
                                promises.push(propTable.getByClass(scopeName, completion.name));
                            }
                        }

                        completions = await methodTable.search(scopeName, keyword);
                        for (const completion of completions) {
                            promises.push(methodTable.getByClass(scopeName, completion.name));
                        }
                    }

                    (await Promise.all(promises)).map((results) => {
                        symbols.push(...results);
                    });
                } else {
                    const promises: Promise<Symbol[]>[] = [];
                    for (const scopeName of scopeNames) {
                        promises.push(propTable.searchAllInClass(scopeName));
                        promises.push(methodTable.searchAllInClass(scopeName));
                    }

                    (await Promise.all(promises)).map((results) => {
                        symbols.push(...results);
                    });
                }

                break;
            case RefKind.ClassTypeDesignator:
                if (ref.type instanceof TypeComposite) {
                    break;
                }

                keyword = ref.type.name;

                if (keyword.length > 0) {
                    const completions = await classTable.search(keyword);
                    for (const completion of completions) {
                        symbols.push(...await classTable.get(completion.name));
                    }
                }

                break;
        }

        return symbols;
    }

    export async function getSignatureHelp(
        phpDoc: PhpDocument, ref: Reference, offset: number
    ): Promise<SignatureHelp | null> {
        const funcTable = App.get<FunctionTable>(FunctionTable);
        const methodTable = App.get<MethodTable>(MethodTable);

        const signatures: SignatureInformation[] = [];
        let activeParameter = 0;
        const symbols: (Function | Method)[] = [];

        if (ref.type instanceof TypeName) {
            if (ref.scope === null) {
                ref.type.resolveReferenceToFqn(phpDoc.importTable);
                symbols.push(...await funcTable.get(ref.type.name));
            } else {
                const classNames: string[] = [];
                if (ref.scope instanceof TypeComposite) {
                    for (const scope of ref.scope.types) {
                        scope.resolveReferenceToFqn(phpDoc.importTable);

                        classNames.push(scope.name);
                    }
                } else {
                    ref.scope.resolveReferenceToFqn(phpDoc.importTable);
                    classNames.push(ref.scope.name);
                }
                for (const className of classNames) {
                    symbols.push(...await methodTable.getByClass(className, ref.type.name));
                }
            }
        }

        if (
            ref.refKind !== RefKind.ArgumentList ||
            ref.ranges === undefined ||
            symbols.length === 0
        ) {
            return null;
        }

        for (const symbol of symbols) {
            if (symbol instanceof Method) {
                signatures.push(Formatter.getMethodSignature(symbol));
            } else if (symbol instanceof Function) {
                signatures.push(Formatter.getFunctionSignature(symbol));
            }
        }

        for (let i = 0; i < ref.ranges.length; i++) {
            if (ref.ranges[i].start <= offset && ref.ranges[i].end >= offset) {
                activeParameter = i;
                break;
            }
        }

        return {
            signatures,
            activeSignature: 0,
            activeParameter,
        };
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
        const methodTable = App.get<MethodTable>(MethodTable);
        const methodInfos: { class: string, method: string }[] = [];

        if (ref.refKind === RefKind.ClassTypeDesignator) {
            const methodName = '__construct';

            ResolveType.forType(ref.type, (type) => {
                methodInfos.push({ class: type.name, method: methodName });
            });
        } else if (
            (ref.refKind === RefKind.Method || ref.refKind === RefKind.MethodCall) &&
            ref.scope !== null
        ) {
            ResolveType.forType(ref.scope, (scope) => {
                ResolveType.forType(ref.type, (type) => {
                    scope.resolveReferenceToFqn(phpDoc.importTable);

                    methodInfos.push({ class: scope.name, method: type.name });
                });
            });
        }

        const methods: Method[] = [];
        const promises: Promise<Method[]>[] = [];

        for (const methodInfo of methodInfos) {
            if (ref.refKind === RefKind.MethodCall) {
                promises.push(methodTable.getByClass(methodInfo.class, methodInfo.method));
            } else if (ref.refKind === RefKind.Method) {
                promises.push(methodTable.getByClass(methodInfo.class, methodInfo.method, (prop) => {
                    return prop.modifier.has(SymbolModifier.STATIC);
                }));
            }
        }

        (await Promise.all(promises)).map((results) => {
            methods.push(...results);
        });

        return methods;
    }

    export async function getPropSymbols(phpDoc: PhpDocument, ref: Reference): Promise<Property[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        const propTable = App.get<PropertyTable>(PropertyTable);
        let classNames: string[] = [];

        if (ref.scope !== null) {
            ResolveType.forType(ref.scope, (type) => {
                type.resolveReferenceToFqn(phpDoc.importTable);

                classNames.push(type.name);
            });
        }

        const properties: Property[] = [];
        const promises: Promise<Property[]>[] = [];

        for (const className of classNames) {
            if (ref.refKind === RefKind.PropertyAccess) {
                promises.push(propTable.getByClass(className, '$' + ref.type.name));
            } else if (ref.refKind === RefKind.Property) {
                promises.push(propTable.getByClass(className, ref.type.name, (prop) => {
                    return prop.modifier.has(SymbolModifier.STATIC);
                }));
            }
        }

        (await Promise.all(promises)).map((results) => {
            properties.push(...results);
        });

        return properties;
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

        if (ref.scope !== null && ref.scope instanceof TypeName) {
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

    export function getReferenceKeyword(ref: Reference, offset: number): string {
        if (ref.type instanceof TypeComposite) {
            return '';
        }

        let keyword = ref.type.name;
        if (
            ref.memberLocation !== undefined &&
            ref.memberLocation.range !== undefined &&
            (
                ref.memberLocation.range.start > offset ||
                ref.memberLocation.range.end < offset
            )
        ) {
            keyword = '';
        }

        return keyword;
    }
}