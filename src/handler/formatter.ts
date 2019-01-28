import { MarkedString, CompletionItem, CompletionItemKind } from "vscode-languageserver";
import { Function } from "../symbol/function/function";
import { TypeName } from "../type/name";
import { PhpDocument } from "../symbol/phpDocument";
import { Class } from "../symbol/class/class";
import { Method } from "../symbol/function/method";
import { SymbolModifier } from "../symbol/meta/modifier";
import { Property } from "../symbol/variable/property";
import { ClassConstant } from "../symbol/constant/classConstant";
import { Reference } from "../symbol/reference";
import { TypeComposite } from "../type/composite";
import { Location } from "../symbol/meta/location";
import { Location as LspLocation, Range as LspRange } from "vscode-languageserver";
import { Range } from "../symbol/meta/range";
import { Constant } from "../symbol/constant/constant";
import { DefineConstant } from "../symbol/constant/defineConstant";
import { Variable } from "../symbol/variable/variable";

export namespace Formatter {
    export function highlightPhp(content: string): MarkedString {
        return {
            language: 'php',
            value: `<?php\n${content}`
        }
    }

    export function types(types: TypeName[]): string {
        return types.map((type): string | null => {
            if (type.isEmptyName()) {
                return null;
            }

            return type.toString();
        }).filter((value) => {
            return value !== null;
        }).join('|');
    }

    export function funcDef(phpDoc: PhpDocument, symbol: Function): MarkedString {
        let params = symbol.parameters.map((param) => {
            return [
                types(param.type.types),
                param.name,
                param.value !== '' ? `= ${param.value}` : null
            ].filter((value) => {
                return value !== null;
            }).join(' ');
        }).join(', ').trim();
        let qualifiedName = symbol.name.getQualified(phpDoc.importTable);

        return highlightPhp(`function ${qualifiedName}(${params})`);
    }

    export function classDef(phpDoc: PhpDocument, symbol: Class): MarkedString {
        let qualifiedName = symbol.name.getQualified(phpDoc.importTable);

        return highlightPhp(`class ${qualifiedName}`);
    }

    export function modifierDef(modifier: SymbolModifier) {
        let modifiers: string[] = [];

        if (modifier.has(SymbolModifier.PUBLIC)) {
            modifiers.push('public');
        }
        if (modifier.has(SymbolModifier.PROTECTED)) {
            modifiers.push('protected');
        }
        if (modifier.has(SymbolModifier.PRIVATE)) {
            modifiers.push('private');
        }
        if (modifier.has(SymbolModifier.STATIC)) {
            modifiers.push('static');
        }
        if (modifier.has(SymbolModifier.ABSTRACT)) {
            modifiers.push('abstract');
        }

        return modifiers.join(' ');
    }

    export function methodDef(phpDoc: PhpDocument, symbol: Method): MarkedString {
        let params = symbol.parameters.map((param) => {
            return [
                types(param.type.types),
                param.name,
                param.value !== '' ? `= ${param.value}` : null
            ].filter((value) => {
                return value !== null;
            }).join(' ');
        }).join(', ').trim();
        let qualifiedName = symbol.name.getQualified(phpDoc.importTable);
        let className = '';
        let modifiers = modifierDef(symbol.modifier);

        if (symbol.scope !== null) {
            className = symbol.scope.getQualified(phpDoc.importTable);
        }

        return highlightPhp(`${modifiers} function ${className}::${qualifiedName}(${params})`);
    }

    export function propDef(phpDoc: PhpDocument, symbol: Property): MarkedString {
        let className = '';
        let modifiers = modifierDef(symbol.modifier);

        if (symbol.scope !== null) {
            className = symbol.scope.getQualified(phpDoc.importTable);
        }

        return highlightPhp(`${modifiers} ${className}::${symbol.name}`);
    }

    export function classConstDef(phpDoc: PhpDocument, symbol: ClassConstant): MarkedString {
        let className = '';

        if (symbol.scope !== null) {
            className = symbol.scope.getQualified(phpDoc.importTable);
        }

        return highlightPhp(`const ${className}::${symbol.name}`);
    }

    export function varRef(phpDoc: PhpDocument, ref: Reference): MarkedString {
        let types: TypeName[] = [];

        if (ref.type instanceof TypeComposite) {
            for (let type of ref.type.types) {
                types.push(type);
            }
        } else {
            types.push(ref.type);
        }

        return highlightPhp(`${Formatter.types(types)} ${ref.refName}`);
    }

    export function constDef(phpDoc: PhpDocument, constant: Constant): MarkedString {
        if (constant instanceof DefineConstant) {
            return highlightPhp(`define(${constant.name.getQualified(phpDoc.importTable)}, ${constant.value})`);
        } else {
            return highlightPhp(`const ${constant.name.getQualified(phpDoc.importTable)} = ${constant.value}`);
        }
    }

    export function toLspLocation(phpDoc: PhpDocument, loc: Location): LspLocation {
        return {
            uri: loc.uri || '',
            range: toLspRange(phpDoc, loc.range || { start: -1, end: -1 })
        }
    }

    export function toLspRange(phpDoc: PhpDocument, range: Range): LspRange {
        return {
            start: phpDoc.getPosition(range.start),
            end: phpDoc.getPosition(range.end)
        };
    }

    export function getFunctionCompletion(phpDoc: PhpDocument, func: Function): CompletionItem {
        return {
            label: func.getName(),
            kind: CompletionItemKind.Function,
            documentation: func.description,
            insertText: func.name.getQualified(phpDoc.importTable)
        };
    }

    export function getClassCompletion(phpDoc: PhpDocument, theClass: Class): CompletionItem {
        return {
            label: theClass.getName(),
            kind: CompletionItemKind.Class,
            documentation: theClass.description,
            insertText: theClass.name.getQualified(phpDoc.importTable)
        };
    }

    export function getConstantCompletion(phpDoc: PhpDocument, constant: Constant): CompletionItem {
        return {
            label: constant.getName(),
            kind: CompletionItemKind.Constant,
            documentation: constant.description,
            insertText: constant.name.getQualified(phpDoc.importTable)
        };
    }

    export function getVariableCompletion(variable: Variable): CompletionItem {
        return {
            label: variable.name,
            kind: CompletionItemKind.Variable,
            documentation: '',
            insertText: variable.name
        };
    }

    export function getPropertyCompletion(prop: Property): CompletionItem {
        let scopeName = '';
        if (prop.scope !== null) {
            scopeName = prop.scope.name;
        }

        return {
            label: `${scopeName}::${prop.name}`,
            kind: CompletionItemKind.Property,
            documentation: prop.description,
            insertText: prop.name
        };
    }

    export function getMethodCompletion(method: Method): CompletionItem {
        let scopeName = '';
        if (method.scope !== null) {
            scopeName = method.scope.name;
        }

        return {
            label: `${scopeName}::${method.getName()}`,
            kind: CompletionItemKind.Method,
            documentation: method.description,
            insertText: method.getName()
        }
    }

    export function getClassConstantCompletion(classConst: ClassConstant): CompletionItem {
        let scopeName = '';
        if (classConst.scope !== null) {
            scopeName = classConst.scope.name;
        }

        return {
            label: `${scopeName}::${classConst.getName()}`,
            kind: CompletionItemKind.Constant,
            documentation: '',
            insertText: classConst.getName()
        }
    }
}