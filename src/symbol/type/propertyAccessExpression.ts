import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { Reference, RefKind, isReference } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TokenKind } from "../../util/parser";
import { TypeComposite } from "../../type/composite";
import { MemberName } from "../name/memberName";

export class PropertyAccessExpression extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.PropertyAccess;

    public type = new TypeName('');
    public location: Location = {};
    public scope: TypeComposite | TypeName = new TypeName('');
    public memberLocation: Location = {};

    private hasArrow = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.Arrow) {
            this.hasArrow = true;

            return true;
        }

        if (!this.hasArrow) {
            if (isReference(other)) {
                this.scope = other.type;
            }
        } else {
            if (other instanceof MemberName) {
                this.type = other.name;
                this.memberLocation = other.location;
            }
        }

        return true;
    }
}