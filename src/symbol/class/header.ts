import { Symbol } from "../symbol";
import { ClassModifier } from "./modifier";
import { NameNode } from "../name/nameNode";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { TreeNode } from "../../util/parseTree";

export class ClassHeader implements Symbol {
    public name: NameNode;
    public modifiers: ClassModifier[];
    public extend: ClassExtend;
    public implement: ClassImplement;

    constructor(public node: TreeNode) {
        this.name = null;
        this.modifiers = [];
        this.extend = null;
        this.implement = null;
    }

    consume(symbol: Symbol) {
        if (symbol instanceof NameNode) {
            this.name = name;
        } else if (symbol instanceof ClassModifier) {
            this.modifiers.push(symbol);
        } else if (symbol instanceof ClassExtend) {
            this.extend = symbol;
        } else if (symbol instanceof ClassImplement) {
            this.implement = symbol;
        }
    }
}