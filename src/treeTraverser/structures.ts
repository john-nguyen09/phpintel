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