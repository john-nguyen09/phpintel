import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { ClassConstant } from "../../symbol/constant/classConstant";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class ClassConstantTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'class_constant',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
    }

    async put(phpDoc: PhpDocument, symbol: ClassConstant) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}