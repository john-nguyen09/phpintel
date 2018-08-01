import { Symbol, TokenSymbol } from "../symbol";
import { Variable } from "./variable";
import { TokenType } from "php7parser";
import { PropertyInitialiser } from "./propertyInitialiser";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";

export class Property extends Variable {
    public modifier: SymbolModifier = null;

    constructor(public node: TreeNode, public doc: PhpDocument) {
        super('');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenType.VariableName) {
                this.name = other.text;
            }
        } else if (other instanceof PropertyInitialiser) {
            this.type = other.expression.type;

            return true;
        }

        return false;
    }

    intake(other: Symbol) {
        if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;
        }
    }
}