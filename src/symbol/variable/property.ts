import { Symbol, TokenSymbol } from "../symbol";
import { Variable } from "./variable";
import { TokenType } from "../../../node_modules/php7parser";
import { PropertyInitialiser } from "./propertyInitialiser";

export class Property extends Variable {
    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenType.VariableName) {
                this.name = other.text;
            }
        } else if (other instanceof PropertyInitialiser) {
            this.type = other.expression.type;
        }

        return false;
    }
}