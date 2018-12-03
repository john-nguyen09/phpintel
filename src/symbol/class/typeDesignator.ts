import { Symbol, Reference, Consumer, NameResolvable } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { ImportTable } from "../../type/importTable";

export class ClassTypeDesignator extends Symbol implements Reference, Consumer, NameResolvable {
    public type: TypeName;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
        }

        return false;
    }

    public resolveName(importTable: ImportTable): void {
        this.type.resolveToFullyQualified(importTable);
    }
}