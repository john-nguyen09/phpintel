import { Symbol, TokenSymbol } from "../symbol";
import { Variable } from "../variable/variable";
import { Expression } from "./expression";
import { TokenType } from "php7parser";

export class Return extends Symbol {
    public returnSymbol: Symbol = null;

    protected expression: Expression;

    consume(other: Symbol) {
        if (
            other instanceof TokenSymbol &&
            (
                other.type == TokenType.Return ||
                other.type == TokenType.Whitespace ||
                other.type == TokenType.Semicolon
            )
        ) {
            return true;
        }

        if (other instanceof Variable) {
            this.returnSymbol = other;
        } else {
            if (!this.expression) {
                this.expression = new Expression(other.node, other.doc);
            }

            this.expression.consume(other);
            this.returnSymbol = this.expression;
        }

        return true;
    }
}