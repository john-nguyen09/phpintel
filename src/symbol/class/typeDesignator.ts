import { Symbol, Reference, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { Name } from "../../type/name";

export class ClassTypeDesignator extends Symbol implements Reference, Consumer {
    public type: Name;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new Name(other.name);
        }

        return false;
    }
}