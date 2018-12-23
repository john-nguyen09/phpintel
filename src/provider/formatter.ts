import { MarkedString } from "vscode-languageserver";
import { Function } from "../symbol/function/function";
import { TypeName } from "../type/name";

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

    export function funcDef(symbol: Function): string {
        let params = symbol.parameters.map((param) => {
            return [
                types(param.type.types),
                param.name,
                param.value !== '' ? `= ${param.value}` : null
            ].filter((value) => {
                return value !== null;
            }).join(' ');
        }).join(', ').trim();

        return `function ${symbol.getName()}(${params})`;
    }
}