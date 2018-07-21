import { Phrase, Token, TokenType, PhraseType } from 'php7parser';
import { Position } from '../symbol/meta/position';
import { Range } from '../symbol/meta/range';

export type TreeNode = Phrase | Token;

export function nodeRange(node: TreeNode, text: string): Range {
    let start = 0;
    let end = 0;

    if (isToken(node)) {
        let t = <Token>node;

        start = t.offset;
        end = t.offset + t.length;
    } else {
        let tFirst = firstToken(node);
        let tLast = lastToken(node);

        if (tFirst && tLast) {
            start = tFirst.offset;
            end = tLast.offset + tLast.length;
        }
    }

    return new Range(
        Position.fromOffset(start, text),
        Position.fromOffset(end, text)
    );
}

export function nodeText(node: TreeNode, text: string): string {
    if (isToken(node)) {
        let t = <Token>node;
        let offset = t.offset;
        let length = t.length;

        return text.slice(offset, offset + length);
    }

    let tFirst = firstToken(node);
    let tLast = lastToken(node);

    if (!tFirst || !tLast) {
        return '';
    }

    let offset = tFirst.offset;
    let endOffset = tLast.offset + tLast.length;

    return text.slice(offset, endOffset);
}

export function firstToken(node: TreeNode): Token {
    if (isToken(node)) {
        return node as Token;
    }

    let t: Token;
    for (let n = 0, l = (<Phrase>node).children.length; n < l; ++n) {
        t = firstToken((<Phrase>node).children[n]);
        if (t !== null) {
            return t;
        }
    }

    return null;
}

export function lastToken(node: TreeNode): Token {
    if (isToken(node)) {
        return node as Token;
    }

    let t: Token;
    for (let n = (<Phrase>node).children.length - 1; n >= 0; --n) {
        t = lastToken((<Phrase>node).children[n]);
        if (t !== null) {
            return t;
        }
    }

    return null;
}

export function isToken(node: Phrase | Token, types?: TokenType[]): boolean {
    return node && (<Token>node).tokenType !== undefined &&
        (!types || types.indexOf((<Token>node).tokenType) > -1);
}

export function isPhrase(node: Phrase | Token, types?: PhraseType[]): boolean {
    return node && (<Phrase>node).phraseType !== undefined &&
        (!types || types.indexOf((<Phrase>node).phraseType) > -1);
}