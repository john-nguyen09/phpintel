import { Symbol, TransformSymbol, Reference, Consumer, Locatable } from "../symbol";
import { QualifiedName } from "../name/qualifiedName";
import { DefineConstant } from "../constant/defineConstant";
import { ArgumentExpressionList } from "../argumentExpressionList";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";

export class FunctionCall extends TransformSymbol implements Consumer, Reference, Locatable {
    public realSymbol: (Symbol & Consumer);
    public type: TypeName;
    public argumentList: ArgumentExpressionList;

    private location: Location;

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.location = new Location(doc.uri, nodeRange(node, doc.text));
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            if (other.name.toLowerCase() == 'define') {
                this.realSymbol = new DefineConstant(this.node, this.doc);
            } else {
                this.type = new TypeName(other.name);
            }

            return true;
        }

        if (this.realSymbol && this.realSymbol != this) {
            return this.realSymbol.consume(other);
        } else if (other instanceof ArgumentExpressionList) {
            this.argumentList = other;

            return true;
        }

        return false;
    }

    getLocation(): Location {
        return this.location;
    }
}