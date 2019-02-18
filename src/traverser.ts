import "reflect-metadata";
import { injectable } from "inversify";
import { TreeNode, isPhrase } from "./util/parseTree";
import { Phrase } from "php7parser";

@injectable()
export class Traverser {
    private visitors: Visitor[] = [];

    traverse(rootNode: TreeNode, visitors: Visitor[]): void {
        this.visitors = visitors;
        this.realTraverse(rootNode, []);
        this.visitors = [];
    }

    private realTraverse(node: TreeNode, spine: Phrase[]) {
        for (let visitor of this.visitors) {
            visitor.preorder(node, spine);
        }

        if (isPhrase(node)) {
            spine.push(node);
            for (let child of node.children) {
                this.realTraverse(child, spine);
            }
            spine.pop();
        }

        for (let visitor of this.visitors) {
            if (visitor.postorder != undefined) {
                visitor.postorder(node);
            }
        }
    }
}

export interface Visitor {
    preorder(node: TreeNode, spine: TreeNode[]): void;
    postorder?(node: TreeNode): void;
}