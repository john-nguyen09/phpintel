import { DbStore, LevelDatasource, SubStore } from "../db";
import { Class } from "../../symbol/class/class";
import { PhpDocument } from "../../symbol/phpDocument";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";
import { CompletionValue, CompletionIndex } from "./index/completionIndex";
import { TypeName } from "../../type/name";
import { Indexer } from "../../index/indexer";
import { App } from "../../app";

@injectable()
export class ClassTable {
    private db: DbStore;
    private nameIndex: DbStore;
    private completionIndex: CompletionIndex;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'class',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.nameIndex = new SubStore(level, {
            name: 'classNameIndex',
            version: 1
        });
        this.completionIndex = new CompletionIndex(level, 'classCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: Class) {
        let name = symbol.getName();

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, name, symbol),
            NameIndex.put(this.nameIndex, phpDoc, name),
            this.completionIndex.put(phpDoc, name)
        ]);
    }

    async get(name: string): Promise<Class[]> {
        let classes: Class[] = [];
        let uris = await NameIndex.get(this.nameIndex, name);

        for (let uri of uris) {
            classes.push(await BelongsToDoc.get<Class>(this.db, uri, name));
        }

        return classes;
    }

    async search(keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword);
    }

    async getByDoc(phpDoc: PhpDocument): Promise<Class[]> {
        const indexer: Indexer = App.get<Indexer>(Indexer);

        if (indexer.isOpen(phpDoc.uri)) {
            return phpDoc.classes;
        }

        return BelongsToDoc.getByDoc<Class>(this.db, phpDoc.uri);
    }

    async removeByDoc(uri: string) {
        let names = await BelongsToDoc.removeByDoc(this.db, uri);

        for (let name of names) {
            await NameIndex.remove(this.nameIndex, uri, name);
            await this.completionIndex.del(uri, name);
        }
    }
}