import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { SymbolModifier } from "../meta/modifier";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";

export class ClassHeader extends Symbol implements Consumer {
    public name: TypeName;
    public modifier: SymbolModifier = new SymbolModifier();
    public extend: ClassExtend;
    public implement: ClassImplement;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenKind.Name:
                    this.name = new TypeName(other.text);
                    break;
                case TokenKind.Abstract:
                case TokenKind.Final:
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