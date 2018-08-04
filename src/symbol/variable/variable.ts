import { Symbol, Consumer, Reference } from "../symbol";
import { Expression } from "../type/expression";
import { Name } from "../../type/name";

export class Variable extends Symbol implements Consumer, Reference {
    public type: Name;

    protected expression: Expression;

    constructor(public name: string, type?: Name) {
        super(null, null);

        if (type) {
            this.type = type;
        }

        this.expression = new Expression(null, null);
    }

    consume(other: Symbol) {
        let result = this.expression.consume(other);

        this.type = this.expression.type;

        return result;
    }
}