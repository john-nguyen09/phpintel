import { Symbol, Consumer, Reference } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { Name } from "../../type/name";

export class TypeDeclaration extends Symbol implements Consumer, Reference {
    public type: Name = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new Name(other.name);

            return true;
        }

        return false;
    }
}