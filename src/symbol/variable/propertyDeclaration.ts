import { CollectionSymbol, Symbol, Consumer } from "../symbol";
import { Property } from "./property";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";

export class PropertyDeclaration extends CollectionSymbol implements Consumer {
    public realSymbols: Symbol[] = [];
    public modifier: SymbolModifier = null;

    consume(other: Symbol): boolean {
        if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;

            return true;
        } else if (other instanceof Property) {
            other.modifier = this.modifier;

            this.realSymbols.push(other);

            return true;
        }

        return false;
    }
}