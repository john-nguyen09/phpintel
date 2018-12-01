import { Symbol, TokenSymbol, ScopeMember } from "../symbol";
import { Variable } from "./variable";
import { PropertyInitialiser } from "./propertyInitialiser";
import { SymbolModifier } from "../meta/modifier";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { TokenKind, PhraseKind } from "../../util/parser";
import { nonenumerable } from "../../util/decorator";
import { TypeComposite } from "../../type/composite";
import { Location } from "../meta/location";

export class Property extends Symbol implements ScopeMember {
    public name: string;
    public location: Location;
    public modifier: SymbolModifier;
    public description: string = '';
    public scope: string = '';

    @nonenumerable
    private _variable: Variable;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this._variable = new Variable('');
        this.location = new Location(doc.uri, nodeRange(node, doc.text));
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            if (other.type == TokenKind.VariableName) {
                this.name = other.text;
            }
        } else if (other instanceof PropertyInitialiser) {
            this._variable.type.push(other.expression.type);

            return true;
        }

        return false;
    }

    public get type(): TypeComposite {
        return this._variable.type;
    }
}