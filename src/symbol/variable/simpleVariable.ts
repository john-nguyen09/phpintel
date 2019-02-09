import { Variable } from "./variable";
import { Symbol, TokenSymbol, ScopeMember } from "../symbol";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Location } from "../meta/location";
import { ScopeVar } from "./scopeVar";
import { Class } from "../class/class";

export class SimpleVariable extends Variable implements FieldGetter, ScopeMember {
    public location: Location;

    public scopeVar: ScopeVar | null = null;

    private scopeClass: Class | null = null;

    constructor() {
        super('');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.VariableName) {
            this.name = other.text;

            if (this.scopeVar !== null) {
                this.type = this.scopeVar.getType(this.name);
            }

            if (this.name == '$this' && this.scopeClass !== null) {
                this.type.push(this.scopeClass.name);
            }
        } else if (other instanceof SimpleVariable) {
            for (let type of other.type.types) {
                this.type.push(type);
            }
        } else {
            return super.consume(other);
        }

        return false;
    }

    getFields(): string[] {
        return ['name', 'type'];
    }

    setScopeClass(scopeClass: Class) {
        this.scopeClass = scopeClass;
    }
}