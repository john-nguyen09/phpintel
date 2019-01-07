import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { TypeName } from "../../type/name";

export class ClassTraitUse extends Symbol implements Consumer {
    public names: TypeName[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedNameList) {
            this.names = other.names.map(name => {
                return new TypeName(name);
            });

            return true;
        }

        return false;
    }
}