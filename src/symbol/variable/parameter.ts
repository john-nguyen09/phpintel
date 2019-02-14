import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TypeDeclaration } from "../type/typeDeclaration";
import { Expression } from "../type/expression";
import { TypeComposite } from "../../type/composite";
import { TokenKind } from "../../util/parser";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";

export class Parameter extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.Parameter;
    public name: string = '';
    public description: string = '';
    public type: TypeComposite = new TypeComposite();
    public location: Location = {};
    public value: string = '';
    public scope: TypeName | null = null;

    protected expression: Expression;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenKind.VariableName:
                    this.name = other.text;
                    break;
                case TokenKind.Equals:
                    this.expression = new Expression();
                    break;
            }
        } else if (other instanceof TypeDeclaration) {
            this.type.push(other.type);

            return true;
        } else if (this.expression != null) {
            this.expression.consume(other);

            if (this.expression.type != null) {
                this.type.push(this.expression.type);
            }

            return true;
        }

        return false;
    }
}