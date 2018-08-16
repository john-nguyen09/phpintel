import { Phrase, Parser } from "php7parser";
import { Consumer, Symbol } from "./symbol";
import { NamespaceDefinition } from "./namespace/definition";
import { ImportTable } from "../type/importTable";
import { NamespaceUse } from "./namespace/Use";
import { nonenumerable } from "../util/decorator";
import { DbStoreInfo } from "../storage/structures";

export class PhpDocument extends Symbol implements Consumer {
    @nonenumerable
    public uri: string;
    @nonenumerable
    public text: string;

    public importTable: ImportTable;
    public symbols: Symbol[] = [];

    constructor(uri: string, text: string) {
        super(null, null);
        this.uri = uri;
        this.text = text;

        this.importTable = new ImportTable();
    }

    getTree(): Phrase {
        return Parser.parse(this.text);
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

        this.symbols.push(other);

        return true;
    }
}