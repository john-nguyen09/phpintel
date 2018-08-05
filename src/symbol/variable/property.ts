import { Symbol, TokenSymbol } from "../symbol";
import { Variable } from "./variable";
import { PropertyInitialiser } from "./propertyInitialiser";
import { SymbolModifier } from "../meta/modifier";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { TokenKind } from "../../util/parser";

export class Property extends Variable {
    public modifier: SymbolModifier;
    public description: string = '';

    constructor(public node: TreeNode, public doc: PhpDocument) {
        super('');
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenKind.VariableName) {
                this.name = other.text;
            }
        } else if (other instanceof PropertyInitialiser) {
            this.type.push(other.expression.type);

            return true;
        }

        return false;
    }
}