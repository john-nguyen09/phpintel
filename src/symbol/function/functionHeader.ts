import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";

export class FunctionHeader extends Symbol implements Consumer {
    public name: TypeName;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol && other.type == TokenKind.Name) {
            this.name = new TypeName(other.text);
        }

        return false;
    }
}