import { Symbol, TransformSymbol } from "./symbol";
import { TreeNode } from "../util/parseTree";
import { QualifiedName } from "./name/qualifiedName";
import { DefineConstant } from "./constant/defineConstant";

export class FunctionCall implements TransformSymbol {
    public realSymbol: Symbol;

    constructor(public node: TreeNode) {
        this.realSymbol = null;
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (other.name.toLowerCase() == 'define') {
                this.realSymbol = new DefineConstant(this.node);

                return true;
            }
        }

        if (this.realSymbol) {
            return this.realSymbol.consume(other);
        }

        return false;
    }
}