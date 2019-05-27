import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";
import { TypeName } from "../../type/name";

export class InterfaceHeader extends Symbol implements Consumer {
    public name: TypeName;

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenKind.Name:
                    this.name = new TypeName(other.text);
            }
        }

        return true;
    }
}