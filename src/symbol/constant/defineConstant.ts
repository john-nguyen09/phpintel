import { Symbol, TokenSymbol, NamedSymbol, Locatable } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { Constant } from "./constant";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Reference } from "../reference";

export class DefineConstant extends Constant implements Reference, FieldGetter, NamedSymbol, Locatable {
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

                return super.consume(args[1]);
            }

            return true;
        }

        return false;
    }
}