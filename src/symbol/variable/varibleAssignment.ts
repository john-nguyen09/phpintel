import { Consumer, Symbol } from "../symbol";
import { SimpleVariable } from "./simpleVariable";

export class VariableAssignment extends Symbol implements Consumer {
    public variable: SimpleVariable;

    consume(other: Symbol): boolean {
        if (other instanceof SimpleVariable) {
            this.variable = other;

            return true;
        } else {
            if (this.variable) {
                this.variable.consume(other);
            }

            return true;
        }
    }
}