import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { Reference, RefKind, isReference } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TypeComposite } from "../../type/composite";
import { TokenKind } from "../../util/parser";
import { MemberName } from "../name/memberName";

export class MethodCallExpression extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.MethodCall;

    public type = new TypeName('');
    public location: Location = {};
    public scope: TypeName | TypeComposite = new TypeName('');

    private hasArrow: boolean = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.Arrow) {
            this.hasArrow = true;
        } else if (!this.hasArrow) {
            if (isReference(other)) {
                this.scope = other.type;
            }
        } else {
            if (other instanceof MemberName) {
                this.type = other.name;
            }
        }

        return true;
    }
}