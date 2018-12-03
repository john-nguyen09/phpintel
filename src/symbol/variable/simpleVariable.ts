import { Variable } from "./variable";
import { Symbol, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Location } from "../meta/location";

export class SimpleVariable extends Variable implements FieldGetter {
    public location: Location;

    constructor() {
        super('');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.VariableName) {
            this.name = other.text;
        } else {
            return super.consume(other);
        }

        return false;
    }

    getFields(): string[] {
        return ['name', 'type'];
    }
}