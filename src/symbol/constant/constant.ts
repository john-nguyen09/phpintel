import { Symbol, TokenSymbol, Consumer, Reference } from "../symbol";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { Expression } from "../type/expression";
import { inspect } from "util";
import { Location } from "../meta/location";
import { PhpDocument } from "../phpDocument";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";

export class Constant extends Symbol implements Consumer, Reference {
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

    inspect(depth: number, options: any) {
        if (depth < 0) {
            return options.stylize(`[${(<any>this).constructor.name}]`, 'special');
        }

        const newObj = {
            name: this.name,
            location: this.location,
            value: this.value,
            type: this.type
        }
        
        const inner = inspect(newObj, options);

        return `${options.stylize((<any>this).constructor.name, 'special')} ${inner}`;
    }
}