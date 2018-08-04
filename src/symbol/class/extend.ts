import { Symbol, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";

export class ClassExtend extends Symbol implements Consumer {
    public name: TypeName = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.name = new TypeName(other.name);

            return true;
        }

        return false;
    }
}