import { Symbol, TokenSymbol, NamedSymbol, Locatable, Reference } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { Constant } from "./constant";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Location } from "../meta/location";

export class DefineConstant extends Symbol implements Reference, FieldGetter, NamedSymbol, Locatable {
    public name: TypeName;
    public location: Location = new Location();

    private constant: Constant = new Constant();

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

    getFields(): string[] {
        return ['name', 'value', 'type'];
    }

    public getName(): string {
        return this.name.toString();
    }
}