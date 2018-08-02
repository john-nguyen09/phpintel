import { Variable } from "./variable";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { Symbol, TokenSymbol } from "../symbol";
import { TokenType } from "php7parser";

export class SimpleVariable extends Variable {
    constructor(public node: TreeNode, public doc: PhpDocument) {
        super('', '');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenType.VariableName) {
                this.name = other.text;
            }
        }

        return false;
    }
}