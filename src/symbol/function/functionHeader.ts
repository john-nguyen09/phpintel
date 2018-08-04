import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TokenType } from "php7parser";
import { Name } from "../../type/name";

export class FunctionHeader extends Symbol implements Consumer {
    public name: Name = null;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol && other.type == TokenType.Name) {
            this.name = new Name(other.text);
        }

        return false;
    }
}