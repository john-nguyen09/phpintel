import { Consumer, Symbol, TokenSymbol, CollectionSymbol } from "../symbol";
import { ClassRef } from "../class/classRef";
import { PropertyRef } from "../variable/propertyRef";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { ScopedMemberName } from "../name/scopedMemberName";

export class PropRefExpression extends CollectionSymbol implements Consumer, Reference {
    public readonly isParentIncluded = true;
    public classRef: ClassRef = new ClassRef();
    public propRef: PropertyRef = new PropertyRef();

    public type = new TypeName('');
    public location = new Location();
    public scope = new TypeName('');

    @nonenumerable
    private hasColonColon = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.propRef.location = new Location(this.location.uri, {
                start: this.location.range.start,
                end: other.node.offset
            });
            this.propRef.scope = this.classRef.type;
            this.scope = this.classRef.type;
        }

        if (!this.hasColonColon) {
            this.classRef.consume(other);
        } else {
            this.propRef.consume(other);
            this.type = this.propRef.type;
        }

        return true;
    }

    get refKind(): RefKind {
        if (!this.type.name.startsWith('$')) {
            return RefKind.ScopedAccess;
        }

        return RefKind.Property;
    }

    get realSymbols() {
        return [
            this.classRef
        ];
    }
}