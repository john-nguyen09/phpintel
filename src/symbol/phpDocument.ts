import { Phrase, Parser } from "php7parser";
import { Consumer, Symbol, needsNameResolve } from "./symbol";
import { NamespaceDefinition } from "./namespace/definition";
import { ImportTable } from "../type/importTable";
import { NamespaceUse } from "./namespace/Use";
import { nonenumerable } from "../util/decorator";
import { Class } from "./class/class";
import { Constant } from "./constant/constant";
import { Function } from "./function/function";
import { ClassConstant } from "./constant/classConstant";
import { Method } from "./function/method";
import { Property } from "./variable/property";
import { isReference, Reference } from "./reference";

export class PhpDocument extends Symbol implements Consumer {
    @nonenumerable
    public text: string;

    @nonenumerable
    private _uri: string;

    public modifiedTime: number = -1;
    public importTable: ImportTable;

    public classes: Class[] = [];
    public functions: Function[] = [];
    public constants: Constant[] = [];
    public classConstants: ClassConstant[] = [];
    public methods: Method[] = [];
    public properties: Property[] = [];
    public references: Reference[] = [];

    constructor(uri: string, text: string) {
        super();

        this._uri = uri;
        this.text = text;
        this.importTable = new ImportTable();
    }

    get uri(): string {
        return this._uri;
    }

    getTree(): Phrase {
        return Parser.parse(this.text);
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

        if (needsNameResolve(other)) {
            other.resolveName(this.importTable);
        }

        return true;
    }

    pushSymbol(symbol: Symbol | null): void {
        if (symbol === null) {
            return;
        }

        if (symbol instanceof Class) {
            this.classes.push(symbol);
        } else if (symbol instanceof Function) {
            this.functions.push(symbol);
        } else if (symbol instanceof Constant) {
            this.constants.push(symbol);
        } else if (symbol instanceof ClassConstant) {
            this.classConstants.push(symbol);
        } else if (symbol instanceof Method) {
            this.methods.push(symbol);
        } else if (symbol instanceof Property) {
            this.properties.push(symbol);
        }
        
        if (isReference(symbol)) {
            this.references.push(symbol);
        }
    }
}