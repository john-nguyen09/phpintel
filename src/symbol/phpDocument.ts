import { Phrase, Parser } from "php7parser";
import { Consumer, Symbol } from "./symbol";
import { NamespaceDefinition } from "./namespace/definition";
import { ImportTable } from "../type/importTable";
import { NamespaceUse } from "./namespace/use";
import { Class } from "./class/class";
import { Constant } from "./constant/constant";
import { Function } from "./function/function";
import { ClassConstant } from "./constant/classConstant";
import { Method } from "./function/method";
import { Property } from "./variable/property";
import { isReference, Reference } from "./reference";
import { Position } from "vscode-languageserver";
import { substr_count } from "../util/string";
import { ScopeVar } from "./variable/scopeVar";
import { ArgumentExpressionList } from "./argumentExpressionList";
import { DefineConstant } from "./constant/defineConstant";
import { GlobalVariable } from "./variable/globalVariable";
import { Range } from "./meta/range";

export class PhpDocument extends Symbol implements Consumer {
    public text: string;

    public uri: string;

    public modifiedTime: number = -1;
    public importTable: ImportTable;

    public classes: Class[];
    public functions: Function[];
    public constants: (Constant | DefineConstant)[];
    public classConstants: ClassConstant[];
    public methods: Method[];
    public properties: Property[];
    public references: Reference[];
    public argumentLists: ArgumentExpressionList[];
    public globalVariables: GlobalVariable[];

    public scopeVarStack: ScopeVar[];

    constructor(uri: string, text: string) {
        super();

        this.uri = uri;
        this.text = text;

        this.refresh();
    }

    refresh() {
        this.importTable = new ImportTable();;

        this.classes = [];
        this.functions = [];
        this.constants = [];
        this.classConstants = [];
        this.methods = [];
        this.properties = [];
        this.references = [];
        this.argumentLists = [];
        this.scopeVarStack = [];
        this.globalVariables = [];
    }

    getTree(): Phrase {
        return Parser.parse(this.text);
    }

    getOffset(line: number, character: number): number {
        let lines = this.text.split('\n');
        let slice = lines.slice(0, line);

        let lineCount = 0;

        if (slice.length > 0) {
            lineCount = slice.map((line) => {
                return line.length;
            }).reduce((total, lineCount) => {
                return total + lineCount;
            });
        }

        return lineCount + slice.length + character;
    }

    getPosition(offset: number): Position {
        let startAt = Math.min(offset, this.text.length);
        let lastNewLine = this.text.lastIndexOf("\n", startAt - 1);
        let character = offset - (lastNewLine + 1);
        let line = offset > 0 ? substr_count(this.text, "\n", 0, offset) : 0;

        return { line, character };
    }

    consume(other: Symbol): boolean {
        if (other instanceof NamespaceDefinition) {
            this.importTable.setNamespace(other.name);

            return true;
        } else if (other instanceof NamespaceUse) {
            for (let alias of other.aliasTable) {
                this.importTable.import(alias.fqn, alias.alias);
            }

            return true;
        }

        return true;
    }

    pushSymbol(symbol: Symbol): void {
        if (symbol instanceof Class) {
            this.classes.push(symbol);
        } else if (symbol instanceof Function) {
            this.functions.push(symbol);
        } else if (symbol instanceof Constant || symbol instanceof DefineConstant) {
            this.constants.push(symbol);
        } else if (symbol instanceof ClassConstant) {
            this.classConstants.push(symbol);
        } else if (symbol instanceof Method) {
            this.methods.push(symbol);
        } else if (symbol instanceof Property) {
            this.properties.push(symbol);
        } else if (symbol instanceof ArgumentExpressionList) {
            this.argumentLists.push(symbol);
        } else if (symbol instanceof GlobalVariable) {
            this.globalVariables.push(symbol);
        }

        if (isReference(symbol)) {
            this.references.push(symbol);
        }
    }

    pushScopeVar(scopeVar: ScopeVar): void {
        this.scopeVarStack.push(scopeVar);
    }

    findRefAt(offset: number): Reference | null {
        let minRef : Reference | null = null;
        let minEnd: number = 0;

        for (const ref of this.references) {
            if (typeof ref.location.range === 'undefined') {
                continue;
            }
            if (ref.location.range.end < offset || ref.location.range.start > offset) {
                continue;
            }

            if (minRef === null) {
                minRef = ref;
                minEnd = ref.location.range.end;
                continue;
            }

            if (ref.location.range.end < minEnd) {
                minRef = ref;
                minEnd = ref.location.range.end;
            }
        }

        return minRef;
    }

    findRefWithin(range: Range, predicate?: (ref: Reference) => boolean): Reference[] {
        const refs: Reference[] = [];

        for (const ref of this.references) {
            if (
                ref.location.range !== undefined &&
                ref.location.range.end <= range.end &&
                ref.scopeRange !== undefined &&
                ref.scopeRange.start === range.start &&
                ref.scopeRange.end === range.end &&
                (typeof predicate === 'undefined' || predicate(ref))
            ) {
                refs.push(ref);
            }
        }

        return refs;
    }

    findScopeVarAt(offset: number): Range | null {
        for (const scopeVar of this.scopeVarStack) {
            if (
                scopeVar.location.range !== undefined &&
                scopeVar.location.range.start <= offset &&
                scopeVar.location.range.end >= offset
            ) {
                return scopeVar.location.range;
            }
        }

        return null;
    }

    findArgumentListAt(offset: number): ArgumentExpressionList | null {
        for (const argumentList of this.argumentLists) {
            if (
                argumentList.location.range !== undefined &&
                argumentList.location.range.start <= offset &&
                argumentList.location.range.end >= offset
            ) {
                return argumentList;
            }
        }

        return null;
    }
}