import { TreeNode, nodeText } from '../util/parseTree';
import { Token, TokenType } from 'php7parser';
import { PhpDocument } from '../phpDocument';
import { nonenumerable } from '../util/decorator';

export abstract class Symbol {
    @nonenumerable
    public node: TreeNode;

    @nonenumerable
    public doc: PhpDocument;

    constructor(node: TreeNode, doc: PhpDocument) {
        this.node = node;
        this.doc = doc;
    }
}

export class TokenSymbol extends Symbol {
    @nonenumerable
    public node: Token;

    public text: string;

    @nonenumerable
    public type: TokenType;

    constructor(token: Token, doc: PhpDocument) {
        super(token, doc);

        this.type = token.tokenType;
        this.text = nodeText(token, doc.text);
    }
}

export abstract class TransformSymbol extends Symbol {
    abstract realSymbol: Symbol;
}

export interface Consumer {
    consume(other: Symbol): boolean;
}

export interface Reference {
    type: string;
    }

export interface ScopeMember {
    scope: string;
}

export function isTransform(symbol: Symbol): symbol is TransformSymbol {
    return symbol != null && 'realSymbol' in symbol;
}

export function isConsumer(symbol: Symbol): symbol is (Symbol & Consumer) {
    return 'consume' in symbol;
}