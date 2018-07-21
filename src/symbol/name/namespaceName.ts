import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";

export class NamespaceName implements Symbol {
    public name: string;

    constructor(public node: TreeNode) {
        this.name = '';
    }

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            this.name += other.text;

            return true;
        }

        return false;
    }
}