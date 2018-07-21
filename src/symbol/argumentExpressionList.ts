import { Symbol, TokenSymbol } from "./symbol";
import { TreeNode } from "../util/parseTree";
import { TokenType } from "../../node_modules/php7parser";

export class ArgumentExpressionList implements Symbol {
    public arguments: Symbol[];

    constructor(public node: TreeNode) {
        this.arguments = [];
    }

    consume(other: Symbol) {
        let isCommaOrWhitespace = false;

        if (
            other instanceof TokenSymbol &&
            (other.type == TokenType.Comma || other.type == TokenType.Whitespace)
        ) {
            isCommaOrWhitespace = true;
        }

        if (!isCommaOrWhitespace) {
            this.arguments.push(other);
        }

        return true;
    }
}