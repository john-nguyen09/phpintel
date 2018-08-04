import { Symbol, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { Name } from "../../type/name";

export class ClassExtend extends Symbol implements Consumer {
    public name: Name = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.name = new Name(other.name);

            return true;
        }

        return false;
    }
}