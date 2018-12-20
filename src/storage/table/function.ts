import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Function } from "../../symbol/function/function";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class FunctionTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'function',
            version: 1
        }, require('../symbolEncoding'));
    }

    async put(phpDoc: PhpDocument, symbol: Function) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}