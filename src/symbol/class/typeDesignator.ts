import { Symbol, Reference, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";

export class ClassTypeDesignator extends Symbol implements Reference, Consumer {
    public type: TypeName;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
        }

        return false;
    }
}