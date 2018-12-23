import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Property } from "../../symbol/variable/property";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";

@injectable()
export class PropertyTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'property',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
    }

    async put(phpDoc: PhpDocument, symbol: Property) {
        return BelongsToDoc.put(this.db, phpDoc, symbol.name, symbol);
    }

    async removeByDoc(uri: string) {
        return BelongsToDoc.removeByDoc(this.db, uri);
    }
}