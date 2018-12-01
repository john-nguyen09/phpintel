import { Symbol, TokenSymbol, Consumer, Reference } from "../symbol";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { Expression } from "../type/expression";
import { Location } from "../meta/location";
import { PhpDocument } from "../phpDocument";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";

export class Constant extends Symbol implements Consumer, Reference, FieldGetter {
    public name: TypeName;
    public expression: Expression;
    public location: Location;

    protected hasEqual: boolean = false;
    protected acceptWhitespace: boolean = true;

    constructor(node: TreeNode | null, doc: PhpDocument | null) {
        super(node, doc);

        if (node != null && doc != null) {
            this.location = new Location(doc.uri, nodeRange(node, doc.text));
        }
    }

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenKind.Name:
                    this.name = new TypeName(other.text);

                    if (this.doc != null) {
                        this.name.resolveToFullyQualified(this.doc.importTable);
                    }

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
                        this.expression = new Expression(other.node, this.doc);
                    }

                    return this.expression.consume(other);
            }

            return true;
        } else {
            if (this.expression == null) {
                this.expression = new Expression(other.node, this.doc);
            }

            this.expression.consume(other);

            return true;
        }
    }

    get value() {
        return this.expression.value;
    }

    get type() {
        return this.expression.type;
    }
    
    getFields(): string[] {
        return ['name', 'value', 'type'];
    }
}