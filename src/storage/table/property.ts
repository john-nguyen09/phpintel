import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Property } from "../../symbol/variable/property";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./nameIndex";

@injectable()
export class PropertyTable {
    public static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'property',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.classIndex = new SubStore(level, {
            name: 'propertyClassIndex',
            version: 1
        });
    }

    async put(phpDoc: PhpDocument, symbol: Property) {
        let className = '';
        if (symbol.scope !== null) {
            className = symbol.scope.getName();
        }

        let key = this.getKey(className, symbol.name);

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key)
        ]);
    }

    async searchByClass(className: string, propName: string): Promise<Property[]> {
        let props: Property[] = [];
        let key = this.getKey(className, propName);
        let uris = await NameIndex.get(this.classIndex, key);

        for (let uri of uris) {
            props.push(await this.db.get(BelongsToDoc.getKey(uri, key)) as Property);
        }

        return props;
    }

    async removeByDoc(uri: string) {
        let props = await BelongsToDoc.removeByDocGetSymbols(this.db, uri) as Property[];

        for (let prop of props) {
            let className = '';
            if (prop.scope !== null) {
                className = prop.scope.getName();
            }

            await NameIndex.remove(this.classIndex, uri, this.getKey(className, prop.name));
        }
    }

    protected getKey(className: string, propName: string): string {
        return `${className}${PropertyTable.CLASS_SEP}${propName}`;
    }
}