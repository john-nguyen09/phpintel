import { Symbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";
import { Variable } from "./variable";
import { Parameter } from "../function/parameter";

export class Scope extends Symbol {
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

    getType(variableName: string) {
        if (variableName in this.variables) {
            return this.variables[variableName].name;
        }

        return '';
    }
}