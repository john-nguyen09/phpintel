import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TypeName } from "../../type/name";
import { ClassTypeDesignator } from "../class/typeDesignator";
import { TokenKind } from "../../util/parser";
import { Location } from "../meta/location";

export class ObjectCreationExpression extends Symbol implements Consumer {
    public location: Location = {};
    public type: TypeName = new TypeName('');
    public scope = null;
    public argumentList: ArgumentExpressionList = new ArgumentExpressionList(this);

    private noOpenParenthesis = 0;
    private startParenthesisOffset = 0;

    public consume(other: Symbol): boolean {
        if (other instanceof ClassTypeDesignator) {
            this.type.name = other.type.name;
        } else if (other instanceof TokenSymbol) {
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

        return true;
    }
}