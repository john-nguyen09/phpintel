import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { QualifiedName } from "../name/qualifiedName";

export class ClassExtend implements Symbol {
    public name: string;

    constructor(public node: TreeNode) {
        this.name = '';
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.name = other.name;

            return true;
        }

        return false;
    }
}