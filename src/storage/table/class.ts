import { DbStore, LevelDatasource, SubStore } from "../db";
import { Class } from "../../symbol/class/class";
import { PhpDocument } from "../../symbol/phpDocument";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./nameIndex";

@injectable()
export class ClassTable {
    private db: DbStore;
    private nameIndex: DbStore;

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
    }

    async put(phpDoc: PhpDocument, symbol: Class) {
        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, symbol.getName(), symbol),
            NameIndex.put(this.nameIndex, phpDoc, symbol.getName())
        ]);
    }

    async get(name: string): Promise<Class[]> {
        let classes: Class[] = [];
        let uris = await NameIndex.get(this.nameIndex, name);

        for (let uri of uris) {
            classes.push(await this.db.get(BelongsToDoc.getKey(uri, name)) as Class);
        }

        return classes;
    }

    async removeByDoc(uri: string) {
        let names = await BelongsToDoc.removeByDoc(this.db, uri);

        for (let name of names) {
            await NameIndex.remove(this.nameIndex, uri, name);
        }
    }
}