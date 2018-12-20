import { Symbol, Reference, Consumer, NameResolvable, Locatable } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { ImportTable } from "../../type/importTable";
import { Location } from "../meta/location";

export class ClassTypeDesignator extends Symbol implements Reference, Locatable, Consumer, NameResolvable {
    public type: TypeName;
    public location: Location;

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