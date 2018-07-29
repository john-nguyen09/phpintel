import { Symbol, Consumer, Reference } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";

export class ConstantAccess extends Symbol implements Consumer, Reference {
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