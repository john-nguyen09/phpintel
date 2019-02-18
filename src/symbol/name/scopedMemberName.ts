import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { Identifier } from "../identifier";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TokenKind } from "../../util/parser";

export class ScopedMemberName extends Symbol implements Consumer {
    public name: TypeName = new TypeName('');
    public location: Location = {};

    consume(other: Symbol): boolean {
        if (other instanceof Identifier) {
            this.name = other.name;
        } else if (other instanceof TokenSymbol && other.type === TokenKind.VariableName) {
            this.name = new TypeName(other.text);
        }

        return true;
    }
}