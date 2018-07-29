import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";

export class ClassTraitUse extends Symbol implements Consumer {
    public names: string[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedNameList) {
            this.names = other.names;

            return true;
        }

        return false;
    }
}