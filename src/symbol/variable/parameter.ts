import { Symbol, TokenSymbol, Reference, Consumer } from "../symbol";
import { TokenType } from "php7parser";
import { TypeDeclaration } from "../type/typeDeclaration";
import { Expression } from "../type/expression";
import { Name } from "../../type/name";

export class Parameter extends Symbol implements Consumer, Reference {
    public name: string = '';
    public type: Name = null;
    public value: string = '';

    protected expression: Expression = null;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenType.VariableName:
                    this.name = other.text;
                    break;
                case TokenType.Equals:
                    this.expression = new Expression(other.node, this.doc);
                    break;
            }
        } else if (other instanceof TypeDeclaration) {
            this.type = other.type;

            return true;
        } else if (this.expression != null) {
            this.expression.consume(other);

            if (this.expression.type != null) {
                this.type = this.expression.type;
            }

            return true;
        }

        return false;
    }
}