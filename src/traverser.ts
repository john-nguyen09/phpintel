import "reflect-metadata";
import { injectable } from "inversify";
import { TreeNode, isPhrase } from "./util/parseTree";

@injectable()
export class Traverser {
    private visitors: Visitor[] = [];

    traverse(rootNode: TreeNode, visitors: Visitor[]): void {
        this.visitors = visitors;
        this.realTraverse(rootNode);
        this.visitors = [];
    }

    private realTraverse(node: TreeNode) {
        for (let visitor of this.visitors) {
            visitor.preorder(node);
        }

        if (isPhrase(node)) {
            for (let child of node.children) {
                this.realTraverse(child);
            }
        }

        for (let visitor of this.visitors) {
            if (visitor.postorder != undefined) {
                visitor.postorder(node);
            }
        }
    }
}

export interface Visitor {
    preorder(node: TreeNode): void;
    postorder?(node: TreeNode): void;
}