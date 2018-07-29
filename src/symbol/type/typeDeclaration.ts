import { Symbol, Consumer, Reference } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";

export class TypeDeclaration extends Symbol implements Consumer, Reference {
    public type: string = '';

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = other.name;

            return true;
        }

        return false;
    }
}