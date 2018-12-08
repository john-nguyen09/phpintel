import { Symbol } from "../symbol";
import { Identifier } from "../identifier";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { TypeName } from "../../type/name";

export class MethodHeader extends Symbol {
    public name: TypeName;
    public modifier: SymbolModifier = new SymbolModifier();

    consume(other: Symbol): boolean {
        if (other instanceof Identifier) {
            this.name = other.name;

            return true;
        } else if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;

            return true;
        }

        return false;
    }
}