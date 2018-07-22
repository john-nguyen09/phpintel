import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode, nodeText } from "../../util/parseTree";
import { ConstantAccess } from "../constant/constantAccess";
import { TokenType } from "../../../node_modules/php7parser";

export class Expression implements Symbol {
    public type: string;
    public value: string;

    constructor(public node: TreeNode) {
        this.type = '';
        this.value = '';
    }

    consume(other: Symbol) {
        this.value = this.getValue(other);
        this.type = this.getType(other);

        return true;
    }

    protected getValue(symbol: Symbol) {
        if (symbol instanceof ConstantAccess) {
            return symbol.value;
        } else if (symbol instanceof TokenSymbol) {
            return symbol.text;
        }
        
        return '';
    }

    protected getType(symbol: Symbol) {
        if (symbol instanceof ConstantAccess) {
            return symbol.type;
        } else if (symbol instanceof TokenSymbol) {
            switch(symbol.type) {
                case TokenType.StringLiteral:
                    return 'string';
                case TokenType.IntegerLiteral:
                    return 'int';
                case TokenType.FloatingLiteral:
                    return 'float';
            }
        }
    }
}