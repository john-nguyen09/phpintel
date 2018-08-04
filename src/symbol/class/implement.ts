import { Symbol, Consumer } from "../symbol";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";

export class ClassImplement extends Symbol implements Consumer {
    public interfaces: TypeName[] = [];

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.interfaces.push(new TypeName(other.name));

            return true;
        } else if (other instanceof QualifiedNameList) {
            this.interfaces.push(...other.names.map(name => {
                return new TypeName(name);
            }));

            return true;
        }

        return false;
    }
}