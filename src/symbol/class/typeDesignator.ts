import { Symbol, Consumer, Locatable } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";

export class ClassTypeDesignator extends Symbol implements Reference, Locatable, Consumer {
    public readonly refKind = RefKind.ClassTypeDesignator;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
        }

        return false;
    }
}