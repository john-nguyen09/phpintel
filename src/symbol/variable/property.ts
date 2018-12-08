import { Symbol, TokenSymbol, ScopeMember, Locatable, DocBlockConsumer } from "../symbol";
import { Variable } from "./variable";
import { PropertyInitialiser } from "./propertyInitialiser";
import { SymbolModifier } from "../meta/modifier";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { TokenKind, PhraseKind } from "../../util/parser";
import { nonenumerable } from "../../util/decorator";
import { TypeComposite } from "../../type/composite";
import { Location } from "../meta/location";
import { TypeName } from "../../type/name";
import { DocBlock } from "../docBlock";

export class Property extends Symbol implements DocBlockConsumer, ScopeMember, Locatable {
    public name: string;
    public location: Location | null = null;
    public modifier: SymbolModifier;
    public description: string = '';
    public scope: TypeName | null = null;

    @nonenumerable
    private _variable: Variable = new Variable('');

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

    consumeDocBlock(doc: DocBlock) {
        let docAst = doc.docAst;

        this.description = docAst.summary;

        this._variable.consumeDocBlock(doc);
    }

    get type(): TypeComposite {
        return this._variable.type;
    }

    set type(val: TypeComposite) {
        this._variable.type = val;
    }
}