import { Symbol, Consumer } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { ScopedMemberName } from "../name/scopedMemberName";

export class ClassConstRef extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.ClassConst;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();
    public scope: TypeName = new TypeName('');

    consume(other: Symbol): boolean {
        if (other instanceof ScopedMemberName) {
            this.type = other.name;
            this.location = other.location;
        }

        return true;
    }
}