import { TreeNode, nodeText } from '../util/parseTree';
import { Token, TokenType } from 'php7parser';
import { PhpDocument } from '../phpDocument';

export interface Symbol {
    node: TreeNode;
    consume(other: Symbol): boolean;
}

export class TokenSymbol implements Symbol {
    public node: TreeNode;
    public text: string;
    public type: TokenType;

    constructor(protected token: Token, doc: PhpDocument) {
        this.node = token;
        this.type = token.tokenType;
        this.text = nodeText(token, doc.text);
    }

    consume(other: Symbol): boolean {
        return false;
    }
}

export abstract class TransformSymbol implements Symbol {
    abstract realSymbol: Symbol;
    abstract node: TreeNode;
    abstract consume(other: Symbol): boolean;
}