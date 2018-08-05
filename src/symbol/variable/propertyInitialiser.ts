import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { Expression } from "../type/expression";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { TokenKind } from "../../util/parser";

export class PropertyInitialiser extends Symbol implements Consumer {
    public expression: Expression;

    protected hasFirstEqual = false;
    protected hasInitialWhitespaces = false;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.expression = new Expression(node, doc);
    }

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