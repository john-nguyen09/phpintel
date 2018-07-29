import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TokenType } from "php7parser";

export class FunctionHeader extends Symbol implements Consumer {
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof TokenSymbol && other.type == TokenType.Name) {
            this.name = other.text;
        }

        return false;
    }
}