import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TypeName } from "../type/name";
import { TokenKind } from "../util/parser";

export class Identifier extends Symbol implements Consumer {
    public name: TypeName = new TypeName('');

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenKind.Name) {
                this.name = new TypeName(other.text);

                return true;
            }
        }

        return false;
    }
}