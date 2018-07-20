import { substr_count } from "../../util/string";

export class Position {
    constructor(public line: number, public character: number) { }

    static fromOffset(offset: number, text: string): Position {
        let textLength = text.length;
        let startAt = Math.max(-(textLength - offset) - 1, -textLength);
        let lastNewLine = text.lastIndexOf("\n", startAt);
        let character = offset - (lastNewLine < 0 ? 0 : lastNewLine + 1);
        let line = offset > 0 ? substr_count(text, "\n", 0, offset) : 0;

        return new Position(line, character);
    }
}