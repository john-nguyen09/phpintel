import { readFile, stat } from "fs";
import { promisify } from "util";
import * as Parser from "tree-sitter";
import * as PhpGrammar from "tree-sitter-php";
import { Position, Range } from "./meta";
import { pathToUri } from "./util/uri";
import { Function } from "./function";
import { Class } from "./class";
import { Constant } from "./constant";
import { Interface } from "./interface";
import { ScopeClass } from "./scope";
import { Method } from "./method";

const readFileAsync = promisify(readFile);
const statAsync = promisify(stat);

export class PhpFile {
    private static parser: Parser | null = null;

    public path: string = '';
    public timeModified: number = 0;
    public lines: string[] = [];

    public scopeClassStack: ScopeClass[] = [];

    public functions: Function[] = [];
    public classes: Class[] = [];
    public constants: Constant[] = [];
    public interfaces: Interface[] = [];

    public static getParser(): Parser {
        if (PhpFile.parser === null) {
            PhpFile.parser = new Parser();
            PhpFile.parser.setLanguage(PhpGrammar);
        }

        return PhpFile.parser;
    }

    public getTextWithin(start: Position, end: Position): string {
        if (
            (end.row < start.row) ||
            (end.row == start.row && end.column <= start.column)
        ) {
            return '';
        }

        if (end.row == start.row) {
            return this.lines[start.row].substr(start.column, end.column - start.column);
        }

        let text = '';

        for (let i = start.row; i <= end.row; i++) {
            if (i == end.row) {
                text += this.lines[i].substr(0, end.column);
                continue;
            }

            if (i == start.row) {
                text += this.lines[i].substr(start.column);
                continue;
            }

            text += this.lines[i];
        }

        return text;
    }

    public parse(): Parser.Tree {
        return PhpFile.getParser().parse(this.text);
    }

    public pushFunction(func: Function) {
        this.functions.push(func);
    }

    public pushConstant(constant: Constant) {
        this.constants.push(constant);
    }

    public pushClass(theClass: Class) {
        this.classes.push(theClass);
    }

    public pushScopeClass(scopeClass: ScopeClass) {
        this.scopeClassStack.push(scopeClass);
    }

    public popScopeClass(): ScopeClass | undefined {
        const scopeClass = this.scopeClassStack.pop();

        if (scopeClass === undefined) {
            console.log('Invalid states: trying to pop an empty scope class stack');
        }

        return scopeClass;
    }

    public pushMethod(method: Method) {
        const scopeClass = this.scopeClass;

        if (scopeClass !== undefined) {
            scopeClass.methods.push(method);

            return;
        }

        console.log(`Uncovered cases: method_declaration is outside of scope classes ${method.name}`);
    }

    get text(): string {
        return this.lines.join('');
    }

    get scopeClass(): ScopeClass | undefined {
        if (this.scopeClassStack.length === 0) {
            return undefined;
        }

        return this.scopeClassStack[this.scopeClassStack.length - 1];
    }

    set text(value: string) {
        this.lines = this.textToLines(value);
    }

    get uri(): string {
        return pathToUri(this.path);
    }

    private textToLines(text: string): string[] {
        const lines: string[] = [];

        let start = 0;
        for (let i = 0; i < text.length; i++) {
            if (text.charAt(i) == '\n') {
                lines.push(text.substr(start, (i - start) + 1));
                start = i + 1;
            }
        }
        lines.push(text.substr(start));

        return lines;
    }
}

export namespace PhpFile {
    export async function create(filePath: string) {
        const stats = await statAsync(filePath);
        const phpFile = new PhpFile();
        phpFile.path = filePath;
        phpFile.timeModified = stats.mtimeMs;
        phpFile.text = (await readFileAsync(filePath)).toString('utf-8');

        return phpFile;
    }
}