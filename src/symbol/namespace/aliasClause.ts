import { Symbol, Consumer, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";

export class NamespaceAliasClause extends Symbol implements Consumer {
    public name: string;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol) {
            switch(other.type) {
                case TokenKind.Name:
                    this.name = other.text;
                    break;
            }
        }

        return false;
    }
}