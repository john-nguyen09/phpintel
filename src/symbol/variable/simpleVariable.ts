import { Variable } from "./variable";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { Symbol, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";

export class SimpleVariable extends Variable {
    constructor(public node: TreeNode, public doc: PhpDocument) {
        super('', undefined);
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenKind.VariableName) {
                this.name = other.text;
            }
        }

        return false;
    }
}