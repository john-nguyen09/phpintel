import { Consumer, Symbol, TokenSymbol } from "../symbol";
import { TokenType } from "php7parser";

export class SymbolModifier implements Consumer {
    static readonly NONE = 0;
    static readonly PUBLIC = 1 << 0;
    static readonly PROTECTED = 1 << 1;
    static readonly PRIVATE = 1 << 2;
    static readonly FINAL = 1 << 3;
    static readonly ABSTRACT = 1 << 4;
    static readonly STATIC = 1 << 5;

    private modifier: number;

    constructor(modifier?: number) {
        if (!modifier) {
            modifier = SymbolModifier.NONE;
        }
        
        this.modifier = modifier;
    }

    has(modifier: number) {
        return (this.modifier & modifier) > 0;
    }

    include(modifier: number) {
        this.modifier |= modifier;
    }

    consume(token: Symbol): boolean {
        if (!(token instanceof TokenSymbol)) {
            return false;
        }

        switch (token.type) {
            case TokenType.Public:
                this.include(SymbolModifier.PUBLIC);
                break;
            case TokenType.Protected:
                this.include(SymbolModifier.PROTECTED);
                break;
            case TokenType.Private:
                this.include(SymbolModifier.PRIVATE);
                break;
            case TokenType.Abstract:
                this.include(SymbolModifier.ABSTRACT);
                break;
            case TokenType.Final:
                this.include(SymbolModifier.FINAL);
                break;
        }

        return true;
    }
}