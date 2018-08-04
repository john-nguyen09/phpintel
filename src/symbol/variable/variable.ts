import { Symbol, Consumer, Reference } from "../symbol";
import { Expression } from "../type/expression";
import { TypeComposite } from "../../type/composite";

export class Variable extends Symbol implements Consumer, Reference {
    public type: TypeComposite = new TypeComposite;

    protected expression: Expression;

    constructor(public name: string, type?: TypeComposite) {
        super(null, null);

        if (type) {
            this.type = type;
        }

        this.expression = new Expression(null, null);
    }

    consume(other: Symbol) {
        let result = this.expression.consume(other);

        this.type.push(this.expression.type);

        return result;
    }
}