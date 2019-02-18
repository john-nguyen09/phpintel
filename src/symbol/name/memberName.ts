import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TokenKind } from "../../util/parser";

export class MemberName extends Symbol implements Consumer {
    public name = new TypeName('');
    public location: Location = {};

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.Name) {
            this.name = new TypeName(other.text);
        }

        return true;
    }
}