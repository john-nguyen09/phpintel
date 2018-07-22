import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";

export class Constant implements Symbol {
    constructor(public node: TreeNode) { }

    consume(other: Symbol) {
        return false;
    }
}