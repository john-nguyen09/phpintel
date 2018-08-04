import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { TokenType } from "php7parser";
import { SymbolModifier } from "../meta/modifier";
import { Name } from "../../type/name";

export class ClassHeader extends Symbol implements Consumer {
    public name: Name = null;
    public modifier: SymbolModifier = new SymbolModifier();
    public extend: ClassExtend = null;
    public implement: ClassImplement = null;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenType.Name:
                    this.name = new Name(other.text);
                    break;
                case TokenType.Abstract:
                case TokenType.Final:
                    this.modifier.consume(other);
                    break;
            }

            return true;
        } else if (other instanceof ClassExtend) {
            this.extend = other;

            return true;
        } else if (other instanceof ClassImplement) {
            this.implement = other;

            return true;
        }

        return false;
    }
}