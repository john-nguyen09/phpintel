import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Property } from "../../symbol/variable/property";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";
import { CompletionIndex, CompletionValue } from "./index/completionIndex";

@injectable()
export class PropertyTable {
    public static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;
    private completionIndex: CompletionIndex;

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
        this.completionIndex = new CompletionIndex(level, 'propertyCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: Property) {
        let className = '';
        if (symbol.scope !== null) {
            className = symbol.scope.name;
        }

        let key = this.getKey(className, symbol.name);

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key),
            this.completionIndex.put(phpDoc, symbol.name, className),
        ]);
    }

    async getByClass(className: string, propName: string): Promise<Property[]> {
        let props: Property[] = [];
        let key = this.getKey(className, propName);
        let uris = await NameIndex.get(this.classIndex, key);

        for (let uri of uris) {
            props.push(await this.db.get(BelongsToDoc.getKey(uri, key)) as Property);
        }

        return props;
    }

    async searchAllInClass(className: string, predicate?: (prop: Property) => boolean): Promise<Property[]> {
        const props: Property[] = [];
        const prefix = this.getKey(className, '');
        const datas = await NameIndex.prefixSearch(this.classIndex, prefix);

        for (let data of datas) {
            const prop = await this.db.get(BelongsToDoc.getKey(data.uri, data.name)) as Property;

            if (typeof predicate !== 'undefined' && !predicate(prop)) {
                continue;
            }

            props.push(prop);
        }

        return props;
    }

    async search(className: string, keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword, className);
    }

    async removeByDoc(uri: string) {
        let props = await BelongsToDoc.removeByDocGetSymbols(this.db, uri) as Property[];

        for (let prop of props) {
            let className = '';
            if (prop.scope !== null) {
                className = prop.scope.name;
            }

            await NameIndex.remove(this.classIndex, uri, this.getKey(className, prop.name));
            await this.completionIndex.del(uri, prop.name, className);
        }
    }

    protected getKey(className: string, propName: string): string {
        return `${className}${PropertyTable.CLASS_SEP}${propName}`;
    }
}