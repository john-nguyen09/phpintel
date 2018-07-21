import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedNameList } from "../list/qualifiedNameList";
import { QualifiedName } from "../name/qualifiedName";

export class ClassImplement implements Symbol {
    public interfaces: string[];

    constructor(public node: TreeNode) {
        this.interfaces = [];
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.interfaces.push(other.name);

            return true;
        } else if (other instanceof QualifiedNameList) {
            this.interfaces.push(...other.names);

            return true;
        }

        return false;
    }
}