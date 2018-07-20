import { Symbol } from "../symbol";
import { Token, TokenType } from "php7parser";
import { SymbolModifier } from "../meta/modifier";
import { TreeNode } from "../../util/parseTree";

export class ClassModifier implements Symbol {
    public node: TreeNode;
    public modifier: number;

    constructor(public token: Token) {
        this.node = token;

        if (token.tokenType === TokenType.Abstract) {
            this.modifier = SymbolModifier.ABSTRACT;
        } else if (token.tokenType === TokenType.Final) {
            this.modifier = SymbolModifier.FINAL;
        }
    }

    consume(other: Symbol) { }
}