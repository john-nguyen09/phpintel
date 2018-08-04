import { Symbol, Consumer, Reference } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { Name } from "../../type/name";

export class ConstantAccess extends Symbol implements Consumer, Reference {
    public static readonly KEYWORD_TYPES: {[key: string]: string} = {
        'false': 'bool',
        'true': 'bool',
        'null': 'null'
    };

    public value: string = '';
    public type: Name = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (ConstantAccess.KEYWORD_TYPES[other.name]) {
                this.value = other.name;
                this.type = new Name(ConstantAccess.KEYWORD_TYPES[other.name]);
            } else {
                this.value = other.name;
                this.type = new Name(other.name);
            }

            return true;
        }

        return false;
    }
}