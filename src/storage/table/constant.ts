import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Constant } from "../../symbol/constant/constant";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";

@injectable()
export class ConstantTable {
    private db: DbStore;
    private nameIndex: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'constant',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.nameIndex = new SubStore(level, {
            name: 'constantNameIndex',
            version: 1
        });
    }

    async put(phpDoc: PhpDocument, symbol: Constant) {
        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol),
            NameIndex.put(this.nameIndex, phpDoc, symbol.getName())
        ]);
    }

    async get(name: string): Promise<Constant[]> {
        let uris = await NameIndex.get(this.nameIndex, name);
        let constSymbols: Constant[] = [];

        for (let uri of uris) {
            constSymbols.push(await this.db.get(BelongsToDoc.getKey(uri, name)));
        }

        return constSymbols;
    }

    async removeByDoc(uri: string) {
        let names = await BelongsToDoc.removeByDoc(this.db, uri);

        for (let name of names) {
            await NameIndex.remove(this.nameIndex, uri, name);
        }
    }
}