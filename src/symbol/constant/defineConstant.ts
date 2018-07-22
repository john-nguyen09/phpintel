import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedName } from "../name/qualifiedName";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TokenType } from "../../../node_modules/php7parser";
import { ConstantAccess } from "./constantAccess";
import { Expression } from "../type/expression";

export class DefineConstant implements Symbol {
    name: string;
    value: string;
    type: string;

    constructor(public node: TreeNode) {
        this.name = '';
        this.value = '';
        this.type = '';
    }

    consume(other: Symbol) {
        if (other instanceof ArgumentExpressionList) {
            if (other.arguments.length == 2) {
                let args = other.arguments;

                if (
                    args[0] instanceof TokenSymbol &&
                    (<TokenSymbol>args[0]).type == TokenType.StringLiteral
                ) {
                    this.name = (<TokenSymbol>args[0]).text.slice(1, -1); // remove quotes
                }

                let expression: Expression = null;

                if (args[1] instanceof Expression) {
                    expression = <Expression>args[1];
                } else {
                    expression = new Expression(args[1].node);
                    expression.consume(args[1]);
                }
                this.value = expression.value;
                this.type = expression.type;
            }

            return true;
        }

        return false;
    }
}