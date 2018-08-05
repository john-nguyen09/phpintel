import { Symbol, Consumer } from "../symbol";
import { Variable } from "./variable";
import { Parameter } from "./parameter";
import { TypeComposite } from "../../type/composite";

export class Scope extends Symbol implements Consumer {
    public variables: { [name: string]: Variable } = {};

    constructor() {
        super(null, null);
    }

    consume(other: Symbol) {
        if (other instanceof Parameter) {
            let variable = new Variable(other.name, other.type);

            this.set(variable);
        }

        return false;
    }

    set(variable: Variable) {
        this.variables[variable.name] = variable;
    }

    getType(variableName: string): TypeComposite {
        if (variableName in this.variables) {
            return this.variables[variableName].type;
        }

        return new TypeComposite();
    }
}