import { Symbol } from "../symbol";
import { NameNode } from "../name/nameNode";
import { TreeNode } from "../../util/parseTree";

export class ClassTraitUse implements Symbol {
    public names: NameNode[];

    constructor(public node: TreeNode) {
        this.names = [];
    }

    consume(other: Symbol) {
        if (other instanceof NameNode) {
            this.names.push(other);
        }
    }
}