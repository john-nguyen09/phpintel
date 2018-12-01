import { Symbol, ScopeMember } from "../symbol";
import { Identifier } from "../identifier";
import { Constant } from "./constant";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { TypeName } from "../../type/name";
import { nonenumerable } from "../../util/decorator";

export class ClassConstant extends Symbol implements ScopeMember {
    public name: TypeName;
    public scope: string = '';

    @nonenumerable
    private constant: Constant;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.constant = new Constant(node, doc);
    }

    consume(other: Symbol) {
        if (other instanceof Identifier) {
            this.name = other.name;

            return true;
        } else {
            this.constant.consume(other);
        }

        return false;
    }

    get value(): string {
        return this.constant.value;
    }

    get type(): TypeName {
        return this.constant.type;
    }
}