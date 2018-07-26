import { Symbol, TokenSymbol } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TokenType } from "php7parser";
import { Constant } from "./constant";

export class DefineConstant extends Constant {
    public name: string = '';

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

                // let expression = new Expression(args[1].node);
                // expression.consume(args[1]);

                // this.value = expression.value;
                // this.type = expression.type;
                super.consume(args[1]);
            }

            return true;
        }

        return false;
    }
}