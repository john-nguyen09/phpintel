import "reflect-metadata";
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

export interface Visitor<T> {
    preorder(node: T): void;
    postorder?(node: T): void;
}

export interface TreeTraverser<T> {
    traverse(rootNode: T, visitors: Visitor<T>[]): void;
}

export function isTraversable<T>(node: T): node is (T & { children: T[] }) {
    return 'children' in node && Array.isArray((<any>node).children);
}