import { Symbol, Consumer } from "../symbol";
import { Location } from "../meta/location";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { SymbolModifier } from "../meta/modifier";
import { ClassTraitUse } from "./traitUse";
import { ClassHeader } from "./header";

export class Class extends Symbol implements Consumer {
    public name: string = '';
    public extend: string = '';
    public location: Location;
    public implements: string[] = [];
    public modifier: SymbolModifier = new SymbolModifier();
    public traits: string[] = [];

    constructor(node: TreeNode, doc: PhpDocument) {
        super(node, doc);

        this.location = new Location(doc.uri, nodeRange(node, doc.text));
    }

    consume(other: Symbol) {
        if (other instanceof ClassHeader) {
            this.name = other.name;
            this.extend = other.extend ? other.extend.name : '';
            this.implements = other.implement ? other.implement.interfaces : [];
            this.modifier = other.modifier;

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