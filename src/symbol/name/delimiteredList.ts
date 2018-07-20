import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { NameNode } from "./nameNode";

export class DelimiteredList extends Symbol {
    public names: NameNode[];

    constructor(node: TreeNode) {
        super(node);

        this.names = [];
    }

    consume(symbol: Symbol) {
        if (symbol instanceof NameNode) {
            this.names.push(symbol);
        }
    }
}