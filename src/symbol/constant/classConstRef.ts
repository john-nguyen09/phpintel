import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { ScopedMemberName } from "../name/scopedMemberName";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";

export class ClassConstRef extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.ClassConst;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();
    public scope: TypeName = new TypeName('');

    @nonenumerable
    private hasWhitespace: boolean = false;

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