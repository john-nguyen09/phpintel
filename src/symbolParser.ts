import { TreeNode, isToken, isPhrase } from "./util/parseTree";
import { Token, Phrase, TokenType } from "../node_modules/php7parser";
import { Symbol } from "./symbol/symbol";
import { NameNode } from "./symbol/name/nameNode";
import { PhpDocument } from "./phpDocument";

export class SymbolParser {
    protected symbolStack: Symbol[] = [];
    protected doc: PhpDocument;

    constructor(doc: PhpDocument) {
        this.doc = doc;
    }

    preorder(node: TreeNode, spine: TreeNode[]) {
        if (isToken(node)) {
            let t = <Token>node;

            switch (t.tokenType) {
                case TokenType.Name:
                    this.symbolStack.push(new NameNode(t, this.doc.text));
                    break;
            }
        } else if (isPhrase(node)) {
            let p = <Phrase>node;
        }
    }
}