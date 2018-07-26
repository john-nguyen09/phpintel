import { Symbol, TokenSymbol } from "../symbol";
import { ClassExtend } from "./extend";
import { ClassImplement } from "./implement";
import { TreeNode } from "../../util/parseTree";
import { TokenType } from "php7parser";
import { SymbolModifier } from "../meta/modifier";
import { PhpDocument } from "../../phpDocument";

export class ClassHeader extends Symbol {
    public name: string = '';
    public modifiers: number[] = [];
    public extend: ClassExtend = null;
    public implement: ClassImplement = null;

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