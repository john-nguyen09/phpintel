import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { TokenType } from "php7parser";
import { QualifiedName } from "../name/qualifiedName";

export class QualifiedNameList implements Symbol {
    public symbols: QualifiedName[];

    constructor(public node: TreeNode) {
        this.symbols = [];
    }

    consume(other: Symbol) {
        if (other instanceof QualifiedName) {
            this.symbols.push(other);

            return true;
        }

        return false;
    }

    get names(): string[] {
        let names: string[] = [];

        for (let symbol of this.symbols) {
            names.push(symbol.name);
        }

        return names;
    }
}