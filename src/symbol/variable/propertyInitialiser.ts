import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { Expression } from "../type/expression";
import { TokenKind } from "../../util/parser";

export class PropertyInitialiser extends Symbol implements Consumer {
    public expression: Expression = new Expression();

    protected hasFirstEqual = false;
    protected hasInitialWhitespaces = false;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenKind.Equals && !this.hasFirstEqual) {
                this.hasFirstEqual = true;
                
                return true;
            } else if (other.type == TokenKind.Whitespace && !this.hasInitialWhitespaces) {
                return true;
            } else {
                this.hasInitialWhitespaces = true;
            }
        } else {
            this.hasInitialWhitespaces = true;
        }

        if (this.hasFirstEqual && this.hasInitialWhitespaces) {
            return this.expression.consume(other);
        }

        return false;
    }
}