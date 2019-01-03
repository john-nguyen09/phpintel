import { Variable } from "./variable";
import { Symbol, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Location } from "../meta/location";
import { nonenumerable } from "../../util/decorator";
import { ScopeVar } from "./scopeVar";

export class SimpleVariable extends Variable implements FieldGetter {
    public location: Location;

    @nonenumerable
    public scopeVar: ScopeVar | null = null;

    constructor() {
        super('');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.VariableName) {
            this.name = other.text;

            if (this.scopeVar !== null) {
                this.type = this.scopeVar.getType(this.name);
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
}