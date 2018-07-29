import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { QualifiedName } from "../name/qualifiedName";

export class ClassImplement extends Symbol implements Consumer {
    public interfaces: string[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.interfaces.push(other.name);

            return true;
        } else if (other instanceof QualifiedNameList) {
            this.interfaces.push(...other.names);

            return true;
        }

        return false;
    }
}