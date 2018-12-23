import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Constant } from "../../symbol/constant/constant";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class ConstantTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'constant',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
    }

    async put(phpDoc: PhpDocument, symbol: Constant) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}