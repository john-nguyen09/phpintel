import { Symbol, Reference, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";

export class ClassTypeDesignator extends Symbol implements Reference, Consumer {
    public type: string;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = other.name;
        }

        return false;
    }
}