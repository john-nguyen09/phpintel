import { Symbol, TokenSymbol } from "../symbol";
import { Identifier } from "../identifier";
import { FunctionHeader } from "./functionHeader";
import { SymbolModifier } from "../meta/modifier";

export class MethodHeader extends FunctionHeader {
    public modifier: SymbolModifier = new SymbolModifier();

    consume(other: Symbol): boolean {
        if (other instanceof Identifier) {
            this.name = other.name;

            return true;
        } else if (other instanceof TokenSymbol) {
            this.modifier.consume(other);

            return true;
        }

        return false;
    }
}