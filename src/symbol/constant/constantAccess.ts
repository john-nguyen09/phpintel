import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { QualifiedName } from "../name/qualifiedName";
import { PhpDocument } from "../../phpDocument";

export class ConstantAccess extends Symbol {
    public static readonly BUILTINS: {[key: string]: string} = {
        'false': 'bool',
        'true': 'bool',
        'null': 'null'
    };

    public value: string = '';
    public type: string = '';

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