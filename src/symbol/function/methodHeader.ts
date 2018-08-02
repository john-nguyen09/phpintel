import { Symbol, TokenSymbol } from "../symbol";
import { Identifier } from "../identifier";
import { FunctionHeader } from "./functionHeader";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";

export class MethodHeader extends FunctionHeader {
    public modifier: SymbolModifier = null;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);
    }

    consume(other: Symbol): boolean {
        if (other instanceof Identifier) {
            this.name = other.name;

            return true;
        } else if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;

            return true;
        }

        return false;
    }
}