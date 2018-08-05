import { Symbol, Consumer, Reference } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";

export class TypeDeclaration extends Symbol implements Consumer, Reference {
    public type: TypeName;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);

            return true;
        }

        return false;
    }
}