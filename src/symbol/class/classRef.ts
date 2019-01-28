import { Symbol, Consumer } from "../symbol";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { QualifiedName } from "../name/qualifiedName";
import { nonenumerable } from "../../util/decorator";
import { ClassConstRefExpression } from "../type/classConstRefExpression";

export class ClassRef extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.Class;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();

    @nonenumerable
    private _scope: TypeName | null = null;

    consume(other: Symbol): boolean {
        if (other instanceof QualifiedName) {
            this.type = new TypeName(other.name);
            this.location = other.location;
        } else if (other instanceof ClassConstRefExpression) {
            this.type = other.scope;
        }

        return true;
    }

    get scope(): TypeName | null {
        return this._scope;
    }

    set scope(value: TypeName | null) {
        this._scope = value;

        if (value !== null && this.type.name === 'self') {
            this.type.name = value.name;
        }
    }
}