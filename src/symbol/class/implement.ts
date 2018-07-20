import { Symbol } from "../symbol";
import { NameNode } from "../name/nameNode";
import { DelimiteredList } from "../name/delimiteredList";
import { TreeNode } from "../../util/parseTree";

export class ClassImplement implements Symbol {
    public interfaces: NameNode[];

    constructor(public node: TreeNode) {
        this.interfaces = [];
    }

    consume(symbol: Symbol) {
        if (symbol instanceof DelimiteredList) {
            this.interfaces.push(...symbol.names);
        }
    }
}