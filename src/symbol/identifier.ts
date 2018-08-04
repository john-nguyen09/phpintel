import { Symbol, TokenSymbol, Consumer } from "./symbol";
import { TokenType } from "php7parser";
import { Name } from "../type/name";

export class Identifier extends Symbol implements Consumer {
    public name: Name = null;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenType.Name) {
                this.name = new Name(other.text);

                return true;
            }
        }

        return false;
    }
}