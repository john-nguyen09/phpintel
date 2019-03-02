import * as Parser from "tree-sitter";
import { Range } from "../meta";

export namespace ParserUtils {
    export function isType(node: Parser.SyntaxNode | null, type: string): node is Parser.SyntaxNode {
        return getType(node) === type;
    }

    export function getType(node: Parser.SyntaxNode | null): string {
        if (node === null) {
            return '';
        }

        return node.type;
    }

    export function getRange(node: Parser.SyntaxNode): Range {
        return {
            start: node.startPosition,
            end: node.endPosition,
        };
    }
}