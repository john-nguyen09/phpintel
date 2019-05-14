import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Function } from "../../symbol/function/function";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";
import { CompletionIndex, CompletionValue } from "./index/completionIndex";
import { TypeName } from "../../type/name";
import { Indexer } from "../../index/indexer";
import { App } from "../../app";

@injectable()
export class FunctionTable {
    private db: DbStore;
    private nameIndex: DbStore;
    private completionIndex: CompletionIndex;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'function',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });

        this.nameIndex = new SubStore(level, {
            name: 'functionNameIndex',
            version: 1
        });

        this.completionIndex = new CompletionIndex(level, 'functionCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: Function) {
        let name = symbol.getName();

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, name, symbol),
            NameIndex.put(this.nameIndex, phpDoc, name),
            this.completionIndex.put(phpDoc, name)
        ]);
    }

    async get(name: string): Promise<Function[]> {
        let funcs: Function[] = [];
        let uris = await NameIndex.get(this.nameIndex, name);

        for (let uri of uris) {
            funcs.push(await BelongsToDoc.get<Function>(this.db, uri, name));
        }

        return funcs;
    }

    async search(keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword);
    }

    async getByDoc(phpDoc: PhpDocument): Promise<Function[]> {
        const indexer: Indexer = App.get<Indexer>(Indexer);

        if (indexer.isOpen(phpDoc.uri)) {
            return phpDoc.functions;
        }

        return BelongsToDoc.getByDoc<Function>(this.db, phpDoc.uri);
    }

    async removeByDoc(uri: string) {
        let names = await BelongsToDoc.removeByDoc(this.db, uri);

        for (let name of names) {
            await NameIndex.remove(this.nameIndex, uri, name);
            await this.completionIndex.del(uri, name);
        }
    }
}