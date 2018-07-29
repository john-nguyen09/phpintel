import { Symbol, TransformSymbol, Reference, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { DefineConstant } from "../constant/defineConstant";
import { ArgumentExpressionList } from "../argumentExpressionList";

export class FunctionCall extends TransformSymbol implements Consumer, Reference {
    public realSymbol: (Symbol & Consumer) = null;
    public type: string;
    public argumentList: ArgumentExpressionList = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (other.name.toLowerCase() == 'define') {
                this.realSymbol = new DefineConstant(this.node, this.doc);
            } else {
                this.type = other.name;
            }

            return true;
        }

        if (this.realSymbol && this.realSymbol != this) {
            return this.realSymbol.consume(other);
        } else if (other instanceof ArgumentExpressionList) {
            this.argumentList = other;

            return true;
        }

        return false;
    }
}