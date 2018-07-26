import { Symbol } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";

export class ClassTraitUse extends Symbol {
    public names: string[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedNameList) {
            this.names = other.names;

            return true;
        }

        return false;
    }
}