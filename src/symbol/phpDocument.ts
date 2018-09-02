import { Phrase, Parser } from "php7parser";
import { Consumer, Symbol } from "./symbol";
import { NamespaceDefinition } from "./namespace/definition";
import { ImportTable } from "../type/importTable";
import { NamespaceUse } from "./namespace/Use";
import { nonenumerable } from "../util/decorator";
import { TextDocument } from "../textDocument";

export class PhpDocument extends Symbol implements Consumer {
    @nonenumerable
    public textDocument: TextDocument;

    @nonenumerable
    private _uri: string;

    public importTable: ImportTable;
    public branchSymbols: Symbol[] = [];
    public symbols: Symbol[] = [];

    constructor(uri: string, text: string) {
        super(null, null);
        this._uri = uri;
        this.textDocument = new TextDocument(text);

        this.importTable = new ImportTable();
    }

    get uri(): string {
        return this._uri;
    }

    get text(): string {
        return this.textDocument.text;
    }

    getTree(): Phrase {
        return Parser.parse(this.textDocument.text);
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

        this.branchSymbols.push(other);

        return true;
    }

    onSymbolDequeued(symbol: Symbol): void {
        this.symbols.push(symbol);
    }
}