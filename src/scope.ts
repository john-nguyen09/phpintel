import { Class } from "./class";
import { Interface } from "./interface";
import { Trait } from "./trait";
import { Variable } from "./variable";

export class ScopeVar {
    private _variables: Map<string, Variable> = new Map<string, Variable>();

    public setVariable(variable: Variable) {

    }
}

export type ScopeClass = Class | Interface | Trait;