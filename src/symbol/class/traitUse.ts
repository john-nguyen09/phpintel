import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { Name } from "../../type/name";

export class ClassTraitUse extends Symbol implements Consumer {
    public names: Name[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedNameList) {
            this.names = other.names.map(name => {
                return new Name(name);
            });

            return true;
        }

        return false;
    }
}