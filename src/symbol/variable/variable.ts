import { Symbol, Consumer, Reference } from "../symbol";
import { Expression } from "../type/expression";
import { TypeComposite } from "../../type/composite";

export class Variable extends Symbol implements Consumer, Reference {
    public type: TypeComposite = new TypeComposite;

    protected expression: Expression;

    constructor(public name: string, type?: TypeComposite) {
        super();

        if (type) {
            this.type = type;
        }

        this.expression = new Expression();
    }

    consume(other: Symbol) {
        let result = this.expression.consume(other);

        if (!this.expression.type.isEmptyName()) {
            this.type.push(this.expression.type);
        }

        return result;
    }
}