import { DbStore, LevelDatasource, SubStore } from "../db";
import { Class } from "../../symbol/class/class";
import { PhpDocument } from "../../symbol/phpDocument";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class ClassTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'class',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
    }

    async put(phpDoc: PhpDocument, symbol: Class) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}