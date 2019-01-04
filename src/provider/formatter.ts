import { MarkedString } from "vscode-languageserver";
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

    export function toLspLocation(phpDoc: PhpDocument, loc: Location): LspLocation {
        return {
            uri: loc.uri,
            range: toLspRange(phpDoc, loc.range)
        }
    }

    export function toLspRange(phpDoc: PhpDocument, range: Range): LspRange {
        return {
            start: phpDoc.getPosition(range.start),
            end: phpDoc.getPosition(range.end)
        };
    }
}