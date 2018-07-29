import { Symbol, Consumer, Reference } from "../symbol";
import { Expression } from "../type/expression";

export class Variable extends Symbol implements Consumer, Reference {
    public type: string;

    protected expression: Expression;

    constructor(public name: string, type?: string) {
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