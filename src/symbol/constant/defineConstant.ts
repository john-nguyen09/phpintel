import { Symbol, TokenSymbol, Locatable } from "../symbol";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { Constant } from "./constant";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Reference, RefKind } from "../reference";
import { Location } from "../meta/location";
import { ImportTable } from "../../type/importTable";
import { TypeComposite } from "../../type/composite";

export class DefineConstant extends Symbol implements FieldGetter, Locatable {
    public readonly refKind = RefKind.Constant;
    public name: TypeName = new TypeName('');
    public description: string;
    public location: Location = {};

    private constant = new Constant();

    consume(other: Symbol) {
        if (other instanceof ArgumentExpressionList) {
            if (other.arguments.length == 2) {
                let args = other.arguments;
                let firstArg = args[0];

                if (
                    firstArg instanceof TokenSymbol &&
                    firstArg.type == TokenKind.StringLiteral
                ) {
                    this.name = new TypeName(firstArg.text.slice(1, -1)); // remove quotes
                }

                return this.constant.consume(args[1]);
            }

            return true;
        }

        return false;
    }

    get value() {
        return this.constant.value;
    }

    get type() {
        return this.constant.type;
    }

    get scope() {
        return this.constant.scope;
    }

    set resolvedType(value: TypeComposite | null) {
        this.constant.resolvedType = value;
    }

    set resolvedValue(value: string) {
        this.constant.resolvedValue = value;
    }

    public getFields(): string[] {
        return ['name', 'value', 'type'];
    }

    public resolveName(importTable: ImportTable): void {
        this.name.resolveDefinitionToFqn(importTable);
    }
}