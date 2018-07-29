import { Symbol, TokenSymbol, TransformSymbol, Consumer, Reference } from "../symbol";
import { ConstantAccess } from "../constant/constantAccess";
import { TokenType } from "php7parser";
import { ClassTypeDesignator } from "../class/typeDesignator";

export class Expression extends TransformSymbol implements Consumer, Reference {
    public realSymbol: Expression = null;

    protected currentSymbol: Symbol;

    consume(other: Symbol): boolean {
        if (other instanceof Expression) {
            this.realSymbol = other;
        }

        if (this.realSymbol == null) {
            if (
                !(other instanceof TokenSymbol) ||
                Expression.hasTokenType(other.type)
            ) {
                this.currentSymbol = other;
            }
        } else {
            return this.realSymbol.consume(other);
        }

        return true;
    }

    get value() {
        if (this.realSymbol) {
            return this.realSymbol.value;
        }

        return this.getValue(this.currentSymbol);
    }

    get type(): string {
        if (this.realSymbol) {
            return this.realSymbol.type;
        }

        return this.getType(this.currentSymbol);
    }

    protected getValue(symbol: Symbol) {
        if (symbol instanceof ConstantAccess) {
            return symbol.value;
        } else if (symbol instanceof TokenSymbol) {
            return symbol.text;
        }
        
        return '';
    }

    protected getType(symbol: Symbol): string {
        if (symbol instanceof ConstantAccess) {
            return symbol.type;
        } else if (symbol instanceof TokenSymbol) {
            let type = Expression.getTokenType(symbol.type);

            if (type) {
                return type;
            }
        } else if (symbol instanceof ClassTypeDesignator) {
            return symbol.type;
        }

        return '';
    }

    static hasTokenType(tokenType: TokenType): boolean {
        return this.getTokenType(tokenType) != null;
    }

    static getTokenType(tokenType: TokenType): string {
        switch(tokenType) {
            case TokenType.StringLiteral:
                return 'string';
            case TokenType.IntegerLiteral:
                return 'int';
            case TokenType.FloatingLiteral:
                return 'float';
        }

        return null;
    }
}