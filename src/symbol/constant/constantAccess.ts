import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedName } from "../name/qualifiedName";

export class ConstantAccess implements Symbol {
    public static readonly BUILTINS: {[key: string]: string} = {
        'false': 'boolean',
        'true': 'boolean',
        'null': 'null'
    };

    public value: string;
    public type: string;

    constructor(public node: TreeNode) {
        this.value = '';
        this.type = '';
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (ConstantAccess.BUILTINS[other.name]) {
                this.value = other.name;
                this.type = ConstantAccess.BUILTINS[other.name];
            } else {
                this.value = other.name;
                this.type = other.name;
            }

            return true;
        }

        return false;
    }
}