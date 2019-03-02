import * as Parser from "tree-sitter";
import { Position } from "./meta";

export class TreeTraverser {
    constructor(private _node: Parser.SyntaxNode) {}

    public setPosition(pos: Position): void {
        this._node = this.find(this._node, pos);
    }

    private find(node: Parser.SyntaxNode, pos: Position): Parser.SyntaxNode {
        for (const child of node.children) {
            if (Position.contains(child.startPosition, child.endPosition, pos)) {
                return this.find(child, pos);
            }
        }

        return node;
    }

    get node(): Parser.SyntaxNode {
        return this._node;
    }

    public parent(): Parser.SyntaxNode | null {
        const parent = this.node.parent;

        if (parent === null) {
            return null;
        }

        this._node = parent;

        return this.node;
    }
}