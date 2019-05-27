import { Symbol, Consumer, ScopeMember } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { QualifiedName } from "../name/qualifiedName";
import { ClassConstRefExpression } from "../type/classConstRefExpression";
import { Class } from "./class";
import { Interface } from "../interface/interface";

export class ClassRef extends Symbol implements Consumer, Reference, ScopeMember {
    public readonly refKind = RefKind.Class;
    public type: TypeName = new TypeName('');
    public location: Location = {};
    public scope: TypeName | null = null;

    consume(other: Symbol): boolean {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
            this.location = other.location;

            if (this.type.name === 'self' && this.scope !== null) {
                this.type = this.scope;
            }
        } else if (other instanceof ClassConstRefExpression) {
            this.type = other.scope;
        }

        return true;
    }

    setScopeClass(scopeClass: Class | Interface) {
        this.scope = scopeClass.name;
    }
}