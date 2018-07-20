import { Symbol } from "../symbol";
import { NameNode } from "../name/nameNode";
import { TreeNode } from "../../util/parseTree";

export class ClassExtend implements Symbol {
    public name: string;

    constructor(public node: TreeNode) {
        this.name = '';
    }

    consume(symbol: Symbol) {
        if (symbol instanceof NameNode) {
            this.name = symbol.name;
        }
    }
}