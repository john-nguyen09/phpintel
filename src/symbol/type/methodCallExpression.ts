import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { Reference, RefKind, isReference } from "../reference";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TypeComposite } from "../../type/composite";
import { TokenKind } from "../../util/parser";
import { MemberName } from "../name/memberName";
import { ArgumentExpressionList } from "../argumentExpressionList";

export class MethodCallExpression extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.MethodCall;

    public type = new TypeName('');
    public location: Location = {};
    public scope: TypeName | TypeComposite = new TypeName('');
    public argumentList: ArgumentExpressionList = new ArgumentExpressionList(this);
    public memberLocation: Location = {};

    private hasArrow: boolean = false;
    private noOpenParenthesis = 0;
    private startParenthesisOffset = 0;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type === TokenKind.Arrow) {
            this.hasArrow = true;
        } else if (!this.hasArrow) {
            if (isReference(other)) {
                this.scope = other.type;
            }
        } else if (other instanceof MemberName) {
            this.type = other.name;
            this.memberLocation = other.location;
        } else {
            if (other instanceof TokenSymbol) {
                if (other.type === TokenKind.OpenParenthesis) {
                    this.noOpenParenthesis++;

                    if (this.noOpenParenthesis === 1) {
                        this.startParenthesisOffset = other.node.offset + other.node.length;
                    }
                } else if (other.type === TokenKind.CloseParenthesis) {
                    this.noOpenParenthesis--;

                    if (this.noOpenParenthesis === 0) {
                        this.argumentList.location = {
                            uri: this.location.uri,
                            range: {
                                start: this.startParenthesisOffset,
                                end: other.node.offset
                            }
                        };
                    }
                }
            } else if (other instanceof ArgumentExpressionList) {
                this.argumentList.arguments = other.arguments;
                this.argumentList.commaOffsets = other.commaOffsets;
            }
        }

        return true;
    }
}