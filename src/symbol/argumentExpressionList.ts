import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TokenKind } from "../util/parser";

export class ArgumentExpressionList extends Symbol implements Consumer {
    public arguments: Symbol[] = [];

    consume(other: Symbol) {
        let isCommaOrWhitespace = false;

        if (
            other instanceof TokenSymbol &&
            (other.type == TokenKind.Comma || other.type == TokenKind.Whitespace)
        ) {
            isCommaOrWhitespace = true;
        }

        if (!isCommaOrWhitespace) {
            this.arguments.push(other);
        }

        return true;
    }
}