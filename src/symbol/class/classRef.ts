import { Symbol, Consumer } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { QualifiedName } from "../name/qualifiedName";

export class ClassRef extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.Class;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();
    public scope: TypeName | null = null;

    consume(other: Symbol): boolean {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
            this.location = other.location;
        }

        return false;
    }
}