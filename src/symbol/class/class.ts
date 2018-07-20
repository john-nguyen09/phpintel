import { Symbol } from "../symbol";
import { Location } from "../meta/location";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";
import { SymbolModifier } from "../meta/modifier";
import { NameNode } from "../name/nameNode";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { ClassModifier } from "./modifier";
import { ClassTraitUse } from "./traitUse";

export class ClassSymbol implements Symbol {
    public name: string;
    public extend: string;
    public location: Location;
    public implements: string[];
    public modifier: SymbolModifier;
    public traits: string[];

    constructor(public node: TreeNode, doc: PhpDocument) {
        this.location = new Location(doc.uri, nodeRange(node, doc.text));
        this.modifier = new SymbolModifier();
        this.implements = [];
    }

    consume(other: Symbol) {
        if (other instanceof NameNode) {
            this.name = other.name;
        } else if (other instanceof ClassExtend) {
            this.extend = other.name;
        } else if (other instanceof ClassImplement) {
            for (let implement of other.interfaces) {
                this.implements.push(implement.name);
            }
        } else if (other instanceof ClassModifier) {
            this.modifier.include(other.modifier);
        } else if (other instanceof ClassTraitUse) {
            for (let trait of other.names) {
                this.traits.push(trait.name);
            }
        }
    }
}