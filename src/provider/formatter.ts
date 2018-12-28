import { MarkedString } from "vscode-languageserver";
import { Function } from "../symbol/function/function";
import { TypeName } from "../type/name";
import { PhpDocument } from "../symbol/phpDocument";

export namespace Formatter {
    export function beautifyPhpContent(content: string): MarkedString {
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

        return beautifyPhpContent(`function ${qualifiedName}(${params})`);
    }
}