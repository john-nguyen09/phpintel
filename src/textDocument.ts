export class TextDocument {
    constructor(public text: string) {
        this.text = text;
    }

    getOffset(line: number, character: number): number {
        let lines = this.text.split('\n');
        let slice = lines.slice(0, line);

        return slice.map((line) => {
            return line.length;
        }).reduce((total, lineCount) => {
            return total + lineCount;
        }) + slice.length + character;
    }
}