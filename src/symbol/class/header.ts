import { Symbol, TokenSymbol } from "../symbol";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { TreeNode } from "../../util/parseTree";
import { TokenType } from "php7parser";
import { SymbolModifier } from "../meta/modifier";

export class ClassHeader implements Symbol {
    public name: string;
    public modifiers: number[];
    public extend: ClassExtend;
    public implement: ClassImplement;

    constructor(public node: TreeNode) {
        this.name = null;
        this.modifiers = [];
        this.extend = null;
        this.implement = null;
    }

    consume(symbol: Symbol) {
        if (symbol instanceof TokenSymbol) {
            switch (symbol.type) {
                case TokenType.Name:
                    this.name = symbol.text;
                    break;
                case TokenType.Abstract:
                    this.modifiers.push(SymbolModifier.ABSTRACT);
                    break;
                case TokenType.Final:
                    this.modifiers.push(SymbolModifier.FINAL);
                    break;
            }

            return true;
        } else if (symbol instanceof ClassExtend) {
            this.extend = symbol;

            return true;
        } else if (symbol instanceof ClassImplement) {
            this.implement = symbol;

            return true;
        }

        return false;
    }
}