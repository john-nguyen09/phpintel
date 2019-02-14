import { Symbol, TransformSymbol, Consumer, Locatable, TokenSymbol } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { DefineConstant } from "../constant/defineConstant";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";

export class FunctionCall extends TransformSymbol implements Consumer, Reference, Locatable, FieldGetter {
    public readonly refKind = RefKind.Function;
    public realSymbol: (Symbol & Consumer);
    public type: TypeName = new TypeName('');
    public argumentList: ArgumentExpressionList = new ArgumentExpressionList();
    public location: Location = {};
    public scope: TypeName | null = null;

    private noOpenParenthesis = 0;
    private startParenthesisOffset = 0;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (other.name.toLowerCase() == 'define') {
                let defineConstant = new DefineConstant();

                defineConstant.location = this.location;
                this.realSymbol = defineConstant;
            } else {
                this.type = new TypeName(other.name);
                this.location = other.location;
            }

            return true;
        }

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
                    this.argumentList.type.name = this.type.name;
                }
            }
        }

        if (this.realSymbol && this.realSymbol != this) {
            return this.realSymbol.consume(other);
        } else if (other instanceof ArgumentExpressionList) {
            this.argumentList.arguments = other.arguments;
            this.argumentList.type.name = this.type.name;
            this.argumentList.commaOffsets = other.commaOffsets;

            return true;
        }

        return false;
    }

    getFields() {
        return [
            'refKind',
            'realSymbol',
            'type',
            'argumentList',
            'location',
            'scope'
        ];
    }
}