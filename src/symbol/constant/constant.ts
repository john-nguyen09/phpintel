import { Symbol, TokenSymbol, Consumer, Reference, NamedSymbol, Locatable } from "../symbol";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { Expression } from "../type/expression";
import { Location } from "../meta/location";
import { PhpDocument } from "../phpDocument";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { ImportTable } from "../../type/importTable";

export class Constant extends Symbol implements Consumer, Reference, FieldGetter, NamedSymbol, Locatable {
    public name: TypeName;
    public expression: Expression;
    public location: Location = new Location();
    public resolvedType: TypeName | null = null;
    public resolvedValue: string | null = null;

    protected hasEqual: boolean = false;
    protected acceptWhitespace: boolean = true;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenKind.Name:
                    this.name = new TypeName(other.text);

                    break;
                case TokenKind.Equals:
                    this.hasEqual = true;
                    break;
                case TokenKind.Whitespace:
                    if (this.expression != null) {
                        this.expression.consume(other);
                    }

                    break;
                default:
                    if (this.expression == null) {
                        this.expression = new Expression();
                    }

                    return this.expression.consume(other);
            }

            return true;
        } else {
            if (this.expression == null) {
                this.expression = new Expression();
            }

            this.expression.consume(other);

            return true;
        }
    }

    get value() {
        if (this.resolvedValue === null) {
            if (typeof this.expression === 'undefined') {
                this.resolvedValue = '';
            } else {
                this.resolvedValue = this.expression.value;
            }
        }

        return this.resolvedValue;
    }

    get type() {
        if (this.resolvedType === null) {
            if (typeof this.expression === 'undefined') {
                this.resolvedType = new TypeName('');
            } else {
                this.resolvedType = this.expression.type;
            }
        }

        return this.resolvedType;
    }
    
    getFields(): string[] {
        return ['name', 'value', 'type'];
    }

    public getName(): string {
        return this.name.toString();
    }

    public resolveName(importTable: ImportTable): void {
        this.name.resolveToFullyQualified(importTable);
    }
}