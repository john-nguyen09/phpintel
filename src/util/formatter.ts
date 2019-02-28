export namespace Formatter {
    export function treeSitterOutput(output: string) {
        let result: string = '';
        let indent: number = 0;

        for (let i = 0; i < output.length; i++) {
            const ch = output.charAt(i);

            if (ch == '(') {
                if (i !== 0) {
                    result += '\n';
                }

                for (let j = 0; j < indent; j++) {
                    result += '\t';
                }

                result += ch;
                indent++;

                continue;
            } else if (ch == ')') {
                result += ch;

                indent--;

                continue;
            } else if (ch == ' ') {
                continue;
            }

            result += ch;
        }

        return result;
    }
}