import { Symbol, Consumer, isScopeMember, NamedSymbol, Locatable } from "../symbol";
import { Location } from "../meta/location";
import { SymbolModifier } from "../meta/modifier";
import { ClassTraitUse } from "./traitUse";
import { ClassHeader } from "./header";
import { TypeName } from "../../type/name";
import { ImportTable } from "../../type/importTable";
import { Property } from "../variable/property";
import { ClassConstRef } from "../constant/classConstRef";
import { PropertyRef } from "../variable/propertyRef";
import { MethodCall } from "../function/methodCall";

export class Class extends Symbol implements Consumer, NamedSymbol, Locatable {
    public name: TypeName;
    public description: string;
    public extend: TypeName | null;
    public location: Location = {};
    public implements: TypeName[] = [];
    public modifier: SymbolModifier = new SymbolModifier();
    public traits: TypeName[] = [];

    consume(other: Symbol) {
        if (other instanceof ClassHeader) {
            this.name = other.name;
            this.extend = other.extend ? other.extend.name : null;
            this.implements = other.implement ? other.implement.interfaces : [];
            this.modifier = other.modifier;

            return true;
        } else if (other instanceof ClassTraitUse) {
            for (let name of other.names) {
                this.traits.push(name);
            }

            return true;
        } else if (
            isScopeMember(other) &&
            !(
                other instanceof ClassConstRef ||
                other instanceof PropertyRef ||
                other instanceof MethodCall
            )
        ) {
            other.scope = this.name;
        }

        return false;
    }

    public getName(): string {
        return this.name.toString();
    }

    public resolveName(importTable: ImportTable): void {
        this.name.resolveDefinitionToFqn(importTable);
    }
}