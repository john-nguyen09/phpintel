import { Symbol } from "../symbol";
import { Function } from "./function";
import { SymbolModifier } from "../meta/modifier";
import { MethodHeader } from "./methodHeader";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { Name } from "../../type/name";
import { nonenumerable } from "../../util/decorator";

export class Method extends Symbol {
    public modifier: SymbolModifier = new SymbolModifier();
    public name: Name = null;

    @nonenumerable
    private func: Function = null;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.func = new Function(node, doc);
    }

    consume(other: Symbol): boolean {
        if (other instanceof MethodHeader) {
            this.modifier = other.modifier;
            this.name = other.name;

            return true;
        }

        return this.func.consume(other);;
    }

    get types() {
        return this.func.types;
    }
}