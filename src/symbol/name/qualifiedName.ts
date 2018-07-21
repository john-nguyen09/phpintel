import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { Phrase } from "../../../node_modules/php7parser";
import { NamespaceName } from "./namespaceName";

export class QualifiedName implements Symbol {
    public name: string;

    constructor(public node: TreeNode) {
    }

    consume(other: Symbol) {
        if (other instanceof NamespaceName) {
            this.name = other.name;

            return true;
        }

        return false;
    }
}