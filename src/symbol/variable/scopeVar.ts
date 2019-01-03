import { Symbol, Consumer } from "../symbol";
import { Variable } from "./variable";
import { Parameter } from "./parameter";
import { TypeComposite } from "../../type/composite";

export class ScopeVar extends Symbol implements Consumer {
    public variables: { [name: string]: TypeComposite } = {};

    consume(other: Symbol) {
        if (other instanceof Parameter) {
            let variable = new Variable(other.name, other.type);

            this.set(variable);
        }

        return false;
    }

    set(variable: Variable) {
        if (!(variable.name in this.variables)) {
            this.variables[variable.name] = variable.type.clone();
            return;
        }

        for (let typeName of variable.type.types) {
            this.variables[variable.name].push(typeName);
        }
    }

    getType(variableName: string): TypeComposite {
        if (variableName in this.variables) {
            let returnType = new TypeComposite();

            for (let type of this.variables[variableName].types) {
                returnType.push(type);
            }
            
            return returnType;
        }

        return new TypeComposite();
    }
}