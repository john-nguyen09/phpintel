import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { QualifiedName } from "../name/qualifiedName";
import { Name } from "../../type/name";

export class ClassImplement extends Symbol implements Consumer {
    public interfaces: Name[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.interfaces.push(new Name(other.name));

            return true;
        } else if (other instanceof QualifiedNameList) {
            this.interfaces.push(...other.names.map(name => {
                return new Name(name);
            }));

            return true;
        }

        return false;
    }
}