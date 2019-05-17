import { Symbol, TokenSymbol, TransformSymbol, Consumer } from "../symbol";
import { ConstantAccess } from "../constant/constantAccess";
import { ClassTypeDesignator } from "../class/typeDesignator";
import { TypeName } from "../../type/name";
import { TokenKind } from "../../util/parser";
import { ObjectCreationExpression } from "./objectCreationExpression";
import { isReference } from "../reference";
import { TypeComposite, ExpressedType } from "../../type/composite";

export class Expression extends TransformSymbol implements Consumer {
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

    get type(): TypeComposite {
        let type: TypeComposite;

        if (this.realSymbol) {
            type = this.realSymbol.type;
        } else {
            type = this.getType(this.currentSymbol);
        }

        if (typeof type === 'undefined') {
            type = new TypeComposite();
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

    protected getType(symbol: Symbol): TypeComposite {
        if (symbol === undefined) {
            return new TypeComposite();
        }

        let type = new TypeComposite();

        if (
            symbol instanceof ConstantAccess ||
            symbol instanceof ClassTypeDesignator ||
            symbol instanceof ObjectCreationExpression
        ) {
            type.push(symbol.type);
        } else if (symbol instanceof TokenSymbol) {
            let tokenType = Expression.getTypeOfToken(symbol.type);

            if (tokenType) {
                type.push(new TypeName(tokenType));
            }
        } else if (isReference(symbol)) {
            const newType = new ExpressedType();
            newType.setReference(symbol);

            type = newType;
        }

        return type;
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