import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TokenType } from "php7parser";
import { TypeName } from "../../type/name";

export class FunctionHeader extends Symbol implements Consumer {
    public name: TypeName = null;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol && other.type == TokenType.Name) {
            this.name = new TypeName(other.text);
        }

        return false;
    }
}