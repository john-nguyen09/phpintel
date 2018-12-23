import { Symbol, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";

export class TypeDeclaration extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.TypeDeclaration;
    public type: TypeName;
    public location: Location = new Location();

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);

            return true;
        }

        return false;
    }
}