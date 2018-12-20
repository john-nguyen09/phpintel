import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Method } from "../../symbol/function/method";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class MethodTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'method',
            version: 1
        }, require('../symbolEncoding'));
    }

    async put(phpDoc: PhpDocument, symbol: Method) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}