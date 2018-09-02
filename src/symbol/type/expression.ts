import { Symbol, TokenSymbol, TransformSymbol, Consumer, Reference } from "../symbol";
import { ConstantAccess } from "../constant/constantAccess";
import { ClassTypeDesignator } from "../class/typeDesignator";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";

export class Expression extends TransformSymbol implements Consumer, Reference {
    public realSymbol: Expression;

    protected currentSymbol: Symbol;

    consume(other: Symbol): boolean {
        if (other instanceof Expression) {
            this.realSymbol = other;

            return true;
        }

        if (this.realSymbol == null) {
            if (
                !(other instanceof TokenSymbol) ||
                Expression.tokenHasType(other.type)
            ) {
                this.currentSymbol = other;
            }
        } else {
            return this.realSymbol.consume(other);
        }

        return true;
    }

    get value(): string {
        if (this.realSymbol) {
            return this.realSymbol.value;
        }

        return this.getValue(this.currentSymbol);
    }

    get type(): TypeName {
        let type: TypeName;

        if (this.realSymbol) {
            type = this.realSymbol.type;
        } else {
            type = this.getType(this.currentSymbol);
        }

        return type;
    }

    protected getValue(symbol: Symbol) {
        if (symbol instanceof ConstantAccess) {
            return symbol.value;
        } else if (symbol instanceof TokenSymbol) {
            return symbol.text;
        }
        
        return '';
    }

    protected getType(symbol: Symbol): TypeName {
        if (
            symbol instanceof ConstantAccess ||
            symbol instanceof ClassTypeDesignator
        ) {
            return symbol.type;
        } else if (symbol instanceof TokenSymbol) {
            let type = Expression.getTypeOfToken(symbol.type);

            if (type) {
                return new TypeName(type);
            }
        }

        return new TypeName('');
    }

    static tokenHasType(tokenType: TokenKind): boolean {
        return this.getTypeOfToken(tokenType) != '';
    }

    static getTypeOfToken(tokenType: TokenKind): string {
        switch(tokenType) {
            case TokenKind.StringLiteral:
                return 'string';
            case TokenKind.IntegerLiteral:
                return 'int';
            case TokenKind.FloatingLiteral:
                return 'float';
        }

        return '';
    }
}