import { Symbol, TokenSymbol } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TokenType } from "php7parser";
import { Constant } from "./constant";
import { PhpDocument } from "../phpDocument";
import { TreeNode } from "../../util/parseTree";
import { Name } from "../../type/name";

export class DefineConstant extends Symbol {
    public name: Name = null;

    private constant: Constant = null;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.constant = new Constant(node, doc);
    }

    consume(other: Symbol) {
        if (other instanceof ArgumentExpressionList) {
            if (other.arguments.length == 2) {
                let args = other.arguments;
                let firstArg = args[0];

                if (
                    firstArg instanceof TokenSymbol &&
                    firstArg.type == TokenType.StringLiteral
                ) {
                    this.name = new Name(firstArg.text.slice(1, -1)); // remove quotes
                }

                this.constant.consume(args[1]);
            }

            return true;
        }

        return false;
    }

    get value(): string {
        return this.constant.value;
    }

    get type(): Name {
        return this.constant.type;
    }
}