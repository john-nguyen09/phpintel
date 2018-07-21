import { Symbol } from "./symbol";
import { TreeNode } from "../util/parseTree";

export class File implements Symbol {
    public node: TreeNode;
    public symbols: Symbol[];

    constructor() {
        this.node = null;
        this.symbols = [];
    }

    consume(other: Symbol) {
        this.symbols.push(other);

        return false;
    }
}