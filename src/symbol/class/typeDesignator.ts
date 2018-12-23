import { Symbol, Consumer, NameResolvable, Locatable } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { ImportTable } from "../../type/importTable";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";

export class ClassTypeDesignator extends Symbol implements Reference, Locatable, Consumer, NameResolvable {
    public readonly refKind = RefKind.ClassTypeDesignator;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();

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