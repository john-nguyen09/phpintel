import { Symbol } from "../symbol";
import { Token } from "php7parser";
import { nodeText, TreeNode } from "../../util/parseTree";

export class NameNode implements Symbol {
    public node: TreeNode;
    public name: string;

    constructor(token: Token, text: string) {
        this.node = token;

        this.name = nodeText(token, text);
    }

    consume(symbol: Symbol) { }
}