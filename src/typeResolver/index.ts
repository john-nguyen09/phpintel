import { Type, KeywordType } from "./type";

export namespace TypeResolver {
    export const KEYWORD_TYPES: Map<string, boolean> = new Map<string, boolean>([
        ['bool', true],
        ['boolean', true],
        ['int', true],
        ['integer', true],
        ['string', true],
        ['real', true],
        ['float', true],
        ['double', true],
        ['object', true],
        ['mixed', true],
        ['array', true],
        ['resource', true],
        ['void', true],
        ['null', true],
        ['scalar', true],
        ['callback', true],
        ['callable', true],
    ]);

    export function toType(type: string): Type {
        if (isKeywordType(type)) {
            return new KeywordType(type);
        }

        return new Type(type);
    }

    export function isKeywordType(type: string) {
        if (type.startsWith('\\')) {
            type = type.substr(1);
        }

        return KEYWORD_TYPES.has(type.toLowerCase());
    }
}