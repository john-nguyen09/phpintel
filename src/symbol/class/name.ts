import { Symbol } from "../symbol";
import { Token } from "php7parser";
import { PhpDocument } from "../../phpDocument";
import { nodeText, TreeNode } from "../../util/parseTree";

export class ClassName implements Symbol {
    public node: TreeNode;
    public text: string;

    constructor(token: Token, doc: PhpDocument) {
        this.node = token;

        this.text = nodeText(token, doc.text);
    }

    consume(other: Symbol) { }
}