import { Consumer, Symbol, TokenSymbol } from "../symbol";
import { SimpleVariable } from "./simpleVariable";
import { TokenKind } from "../../util/parser";

export class VariableAssignment extends Symbol implements Consumer {
    public variable: SimpleVariable;

    private hasEqual = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.Equals) {
            this.hasEqual = true;

            return true;
        }

        if (other instanceof SimpleVariable && !this.hasEqual) {
            this.variable = other;

            return true;
        } else {
            if (this.variable) {
                this.variable.consume(other);
            }

            return true;
        }
    }
}