import { Symbol, Consumer } from "../symbol";
import { Location } from "../meta/location";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { SymbolModifier } from "../meta/modifier";
import { ClassTraitUse } from "./traitUse";
import { ClassHeader } from "./header";
import { TypeName } from "../../type/name";

export class Class extends Symbol implements Consumer {
    public name: TypeName;
    public extend: TypeName | null;
    public location: Location;
    public implements: TypeName[] = [];
    public modifier: SymbolModifier = new SymbolModifier();
    public traits: TypeName[] = [];

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.location = new Location(doc.uri, nodeRange(node, doc.text));
    }

    consume(other: Symbol) {
        if (other instanceof ClassHeader) {
            this.name = other.name;
            this.extend = other.extend ? other.extend.name : null;
            this.implements = other.implement ? other.implement.interfaces : [];
            this.modifier = other.modifier;

            if (this.doc != null) {
                this.name.resolveToFullyQualified(this.doc.importTable);
            }

            return true;
        } else if (other instanceof ClassTraitUse) {
            for (let name of other.names) {
                this.traits.push(name);
            }

            return true;
        }

        return false;
    }
}