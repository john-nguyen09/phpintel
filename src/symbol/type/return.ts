import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { Variable } from "../variable/variable";
import { Expression } from "./expression";
import { TokenKind } from "../../util/parser";

export class Return extends Symbol implements Consumer {
    public returnSymbol: Symbol;

    protected expression: Expression;

    consume(other: Symbol) {
        if (
            other instanceof TokenSymbol &&
            (
                other.type == TokenKind.Return ||
                other.type == TokenKind.Whitespace ||
                other.type == TokenKind.Semicolon
            )
        ) {
            return true;
        }

        if (other instanceof Variable) {
            this.returnSymbol = other;
        } else {
            if (!this.expression) {
                this.expression = new Expression();
            }

            this.expression.consume(other);
            this.returnSymbol = this.expression;
        }

        return false;
    }
}