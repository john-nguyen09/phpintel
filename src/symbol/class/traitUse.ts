import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedNameList } from "../list/qualifiedNameList";

export class ClassTraitUse implements Symbol {
    public names: string[];

    constructor(public node: TreeNode) {
        this.names = [];
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedNameList) {
            this.names = other.names;

            return true;
        }

        return false;
    }
}