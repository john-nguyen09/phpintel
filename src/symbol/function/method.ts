import { Symbol } from "../symbol";
import { Function } from "./function";
import { SymbolModifier } from "../meta/modifier";
import { MethodHeader } from "./methodHeader";

export class Method extends Function {
    public modifier: SymbolModifier = new SymbolModifier();

    consume(other: Symbol): boolean {
        let result = super.consume(other);

        if (other instanceof MethodHeader) {
            this.modifier = other.modifier;
        }

        return result;
    }
}