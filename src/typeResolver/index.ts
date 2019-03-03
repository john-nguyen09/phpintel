import { Type, KeywordType } from "./type";
import * as Parser from "tree-sitter";

export namespace TypeResolver {
    export const KEYWORD_TYPES: Map<string, string> = new Map<string, string>([
        ['bool', 'bool'],
        ['boolean', 'bool'],
        ['true', 'bool'],
        ['false', 'bool'],
        ['int', 'int'],
        ['integer', 'int'],
        ['string', 'string'],
        ['real', 'real'],
        ['float', 'float'],
        ['double', 'double'],
        ['object', 'object'],
        ['mixed', 'mixed'],
        ['array', 'array'],
        // ['resource', 'resource'],
        ['void', 'void'],
        ['null', 'null'],
        // ['scalar', 'scalar'],
        // ['callback', 'callback'],
        // ['callable', 'callable'],
    ]);

    export function stringToType(type: string): Type {
        if (isKeywordType(type)) {
            return new KeywordType(type);
        }

        return new Type(type);
    }

    export function getNodeType(node: Parser.SyntaxNode): Type {
        if (node.type === 'integer') {
            return new KeywordType('int');
        }

        if (node.type === 'string') {
            return new KeywordType('string');
        }

        if (node.type === 'qualified_name' || node.type === 'name') {
            const keywordTypeString = KEYWORD_TYPES.get(node.text);

            if (keywordTypeString !== undefined) {
                return new KeywordType(keywordTypeString);
            }
        }

        if (node.type === 'object_creation_expression') {
            const nameNode = node.firstNamedChild;

            if (nameNode !== null) {
                return new Type(nameNode.text);
            }
        }

        return new Type(''); // Cannot determine the type
    }

    export function isKeywordType(type: string) {
        if (type.startsWith('\\')) {
            type = type.substr(1);
        }

        return KEYWORD_TYPES.has(type.toLowerCase());
    }
}