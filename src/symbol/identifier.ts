import { Symbol, TokenSymbol } from "./symbol";
import { TokenType } from "php7parser";

export class Identifier extends Symbol {
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenType.Name) {
                this.name = other.text;

                return true;
            }
        }

        return false;
    }
}