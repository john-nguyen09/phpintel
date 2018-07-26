import { Symbol, TokenSymbol } from "../symbol";
import { TreeNode } from "../../util/parseTree";
import { PhpDocument } from "../../phpDocument";
import { TokenType } from "php7parser";
import { TypeDeclaration } from "../type/typeDeclaration";

export class Parameter extends Symbol {
    public type: string = '';
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            switch (other.type) {
                case TokenType.VariableName:
                    this.name = other.text;
                    break;
            }
        } else if (other instanceof TypeDeclaration) {
            this.type = other.type;

            return true;
        }

        return false;
    }
}