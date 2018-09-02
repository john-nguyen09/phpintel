import { TreeTraverser, Visitor, isTraversable} from "./structures";
import { injectable } from "inversify";

@injectable()
export class RecursiveTraverser<T> implements TreeTraverser<T> {
    private visitors: Visitor<T>[] = [];

    traverse(rootNode: T, visitors: Visitor<T>[]): void {
        this.visitors = visitors;
        this.realTraverse(rootNode);
        this.visitors = [];
    }

    private realTraverse(node: T) {
        for (let visitor of this.visitors) {
            visitor.preorder(node);
        }

        if (isTraversable<T>(node)) {
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