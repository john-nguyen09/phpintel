import { Class } from "./class";
import { Interface } from "./interface";
import { Trait } from "./trait";
import { Variable } from "./variable";

export class ScopeVar {
    private _variables: Map<string, Variable> = new Map<string, Variable>();

    public setVariable(variable: Variable) {
        const currentVariable = this._variables.get(variable.name);

        if (currentVariable === undefined) {
            this._variables.set(variable.name, variable);

            return;
        }

        currentVariable.type.push(variable.type);
    }

    public getVariable(name: string): Variable | undefined {
        return this._variables.get(name);
    }
}

export type ScopeClass = Class | Interface | Trait;