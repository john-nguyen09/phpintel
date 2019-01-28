import { Symbol, DocBlockConsumer, ScopeMember, NamedSymbol, Locatable } from "../symbol";
import { Function } from "./function";
import { SymbolModifier } from "../meta/modifier";
import { MethodHeader } from "./methodHeader";
import { TypeName } from "../../type/name";
import { nonenumerable } from "../../util/decorator";
import { DocBlock } from "../docBlock";
import { Location } from "../meta/location";
import { Variable } from "../variable/variable";

export class Method extends Symbol implements DocBlockConsumer, ScopeMember, NamedSymbol, Locatable {
    public modifier: SymbolModifier = new SymbolModifier();
    public name: TypeName;
    public location: Location = {};
    public description: string = '';
    public scope: TypeName | null = null;

    @nonenumerable
    private func: Function = new Function();

    consume(other: Symbol): boolean {
        if (other instanceof MethodHeader) {
            this.modifier = other.modifier;
            this.name = other.name;

            return true;
        }

        return this.func.consume(other);;
    }

    consumeDocBlock(doc: DocBlock) {
        this.func.consumeDocBlock(doc);

        this.description = this.func.description;
    }

    get types() {
        return this.func.types;
    }

    pushType(type: TypeName | null) {
        if (type === null) {
            return;
        }

        this.func.typeAggregate.push(type);
    }

    get variables() {
        return this.func.scopeVar.variables;
    }

    setVariable(variable: Variable) {
        this.func.scopeVar.set(variable);
    }

    get parameters() {
        return this.func.parameters;
    }

    public getName(): string {
        return this.name.toString();
    }
}