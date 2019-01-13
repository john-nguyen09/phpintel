import { CollectionSymbol, Consumer, Symbol, TokenSymbol } from "../symbol";
import { ClassRef } from "../class/classRef";
import { ClassConstRef } from "../constant/classConstRef";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";

export class ClassConstRefExpression extends CollectionSymbol implements Consumer, Reference {
    public isParentIncluded = true;
    public readonly refKind = RefKind.ScopedAccess;
    public type: TypeName = new TypeName('');
    public location: Location = new Location();
    public scope: TypeName = new TypeName('');

    public classRef: ClassRef = new ClassRef();
    public classConstRef: ClassConstRef = new ClassConstRef();

    @nonenumerable
    private hasColonColon: boolean = false;

    consume(other: Symbol): boolean {
        let results = false;

        if (other instanceof TokenSymbol && other.type === TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.classConstRef.scope = this.classRef.type;

            return true;
        }

        if (!this.hasColonColon) {
            results = this.classRef.consume(other);
            this.scope = this.classRef.type;
        } else {
            results = this.classConstRef.consume(other);
            this.type = this.classConstRef.type;
        }

        return results;
    }

    get realSymbols(): Symbol[] {
        return [
            this.classRef,
            this.classConstRef
        ]
    }
}