import { Symbol, Consumer } from "../symbol";
import { FunctionHeader } from "./functionHeader";
import { Parameter } from "../variable/parameter";
import { Scope } from "../variable/scope";
import { Return } from "../type/return";
import { Variable } from "../variable/variable";
import { Expression } from "../type/expression";
import { SimpleVariable } from "../variable/simpleVariable";
import { TypeAggregate } from "../../type/aggregate";

export class Function extends Symbol implements Consumer {
    public name: string = '';
    public parameters: Parameter[] = [];
    public scopeVar: Scope = new Scope();
    public typeAggregate: TypeAggregate = new TypeAggregate();

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
                this.typeAggregate.push(this.scopeVar.getType(returnSymbol.name));
            } else if (returnSymbol instanceof Expression) {
                this.typeAggregate.push(returnSymbol.type);
            }

            return true;
        } else if (other instanceof SimpleVariable) {
            this.scopeVar.set(other);

            return true;
        }

        return false;
    }

    get types(): string[] {
        return this.typeAggregate.types;
    }
}