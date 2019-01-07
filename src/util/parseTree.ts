import { Phrase, Token } from 'php7parser';
import { Range } from '../symbol/meta/range';

export type TreeNode = Phrase | Token;

export function nodeRange(node: TreeNode, text: string): Range {
    let start = 0;
    let end = 0;

    if (isToken(node)) {
        let t = node;

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
        start,
        end
    );
}

export function nodeText(node: TreeNode, text: string): string {
    if (isToken(node)) {
        let t = node;
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

export function firstToken(node: TreeNode): Token | null {
    if (isToken(node)) {
        return node;
    }

    let t: Token | null;
    for (let n = 0, l = node.children.length; n < l; ++n) {
        t = firstToken(node.children[n]);
        if (t !== null) {
            return t;
        }
    }

    return null;
}

export function lastToken(node: TreeNode): Token | null {
    if (isToken(node)) {
        return node;
    }

    let t: Token | null;
    for (let n = node.children.length - 1; n >= 0; --n) {
        t = lastToken(node.children[n]);
        if (t !== null) {
            return t;
        }
    }

    return null;
}

export function isToken(node: Phrase | Token): node is Token {
    return node && (<Token>node).tokenType !== undefined;
}

export function isPhrase(node: Phrase | Token): node is Phrase {
    return node && (<Phrase>node).phraseType !== undefined;
}