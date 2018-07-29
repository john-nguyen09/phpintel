import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TokenType } from "php7parser";

export class ArgumentExpressionList extends Symbol implements Consumer {
    public arguments: Symbol[] = [];

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