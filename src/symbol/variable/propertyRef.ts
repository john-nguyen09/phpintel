import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { ScopedMemberName } from "../name/scopedMemberName";
import { TokenKind } from "../../util/parser";

export class PropertyRef extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.Property;
    public type: TypeName = new TypeName('');
    public location: Location = {};
    public scope: TypeName = new TypeName('');

    private hasWhitespace = false;

    consume(other: Symbol): boolean {
        if (this.hasWhitespace) {
            return true;
        }

        if (other instanceof ScopedMemberName) {
            this.type = other.name;
            this.location = other.location;
        } else if (other instanceof TokenSymbol && other.type === TokenKind.Whitespace) {
            this.hasWhitespace = true;
        }

        return true;
    }
}