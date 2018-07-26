import { Symbol } from "../symbol";
import { Parameter } from "./parameter";
import { FunctionHeader } from "./header";
import { Scope } from "../variable/scope";
import { Return } from "../type/return";
import { Variable } from "../variable/variable";
import { Expression } from "../type/expression";
import { SimpleVariable } from "../variable/simpleVariable";

export class Function extends Symbol {
    public name: string = '';
    public parameters: Parameter[] = [];
    public scopeVar: Scope = new Scope();
    public types: string[] = [];

    consume(other: Symbol) {
        if (other instanceof Parameter) {
            this.parameters.push(other);
            this.scopeVar.consume(other);

            return true;
        } else if (other instanceof FunctionHeader) {
            this.name = other.name;

            return true;
        } else if (other instanceof Return) {
            let returnSymbol = other.returnSymbol;

            if (returnSymbol instanceof Variable) {
                this.types.push(this.scopeVar.getType(returnSymbol.name));
            } else if (returnSymbol instanceof Expression) {
                this.types.push(returnSymbol.type);
            }

            return true;
        } else if (other instanceof SimpleVariable) {
            this.scopeVar.set(other);

            return true;
        }

        return false;
    }
}