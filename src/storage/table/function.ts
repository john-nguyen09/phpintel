import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Function } from "../../symbol/function/function";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./nameIndex";

@injectable()
export class FunctionTable {
    private db: DbStore;
    private nameIndex: DbStore;

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
    }

    async put(phpDoc: PhpDocument, symbol: Function) {
        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol),
            NameIndex.put(this.nameIndex, phpDoc, symbol.getName())
        ]);
    }

    async get(name: string): Promise<Function[]> {
        let funcs: Function[] = [];
        let uris = await NameIndex.get(this.nameIndex, name);

        for (let uri of uris) {
            funcs.push(await this.db.get(BelongsToDoc.getKey(uri, name)) as Function);
        }

        return funcs;
    }

    async removeByDoc(uri: string) {
        let names = await BelongsToDoc.removeByDoc(this.db, uri);

        for (let name of names) {
            await NameIndex.remove(this.nameIndex, uri, name);
        }
    }
}