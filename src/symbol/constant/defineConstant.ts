import { Symbol, TokenSymbol } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { Constant } from "./constant";
import { PhpDocument } from "../phpDocument";
import { TreeNode } from "../../util/parseTree";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";

export class DefineConstant extends Symbol {
    public name: TypeName;

    private constant: Constant;

    constructor(node: TreeNode | null, doc: PhpDocument | null) {
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
                    firstArg.type == TokenKind.StringLiteral
                ) {
                    this.name = new TypeName(firstArg.text.slice(1, -1)); // remove quotes
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

    get type(): TypeName {
        return this.constant.type;
    }
}