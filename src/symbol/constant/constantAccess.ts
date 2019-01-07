import { Symbol, Consumer } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";

export class ConstantAccess extends Symbol implements Consumer, Reference {
    public readonly refKind = RefKind.ConstantAccess;
    public static readonly KEYWORD_TYPES: {[key: string]: string} = {
        'false': 'bool',
        'true': 'bool',
        'null': 'null'
    };

    public value: string = '';
    public type: TypeName = new TypeName('');
    public location: Location = new Location();
    public scope: TypeName | null = null;

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (ConstantAccess.KEYWORD_TYPES[other.name]) {
                this.value = other.name;
                this.type = new TypeName(ConstantAccess.KEYWORD_TYPES[other.name]);
            } else {
                this.value = other.name;
                this.type = new TypeName(other.name);
            }

            return true;
        }

        return false;
    }
}