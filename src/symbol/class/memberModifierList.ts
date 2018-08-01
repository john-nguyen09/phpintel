import { Consumer, Symbol, TokenSymbol } from "../symbol";
import { SymbolModifier } from "../meta/modifier";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";

export class MemberModifierList extends Symbol implements Consumer {
    public modifier: SymbolModifier = new SymbolModifier();

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.modifier.include(SymbolModifier.PUBLIC);
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            this.modifier.consume(other);
        }

        return false;
    }
}