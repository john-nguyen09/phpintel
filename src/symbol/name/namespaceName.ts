import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";

export class NamespaceName extends Symbol {
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            this.name += other.text;

            return true;
        }

        return false;
    }
}