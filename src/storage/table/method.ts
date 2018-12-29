import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Method } from "../../symbol/function/method";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./nameIndex";

@injectable()
export class MethodTable {
    public static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'method',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.classIndex = new SubStore(level, {
            name: 'methodClassIndex',
            version: 1
        });
    }

    async put(phpDoc: PhpDocument, symbol: Method) {
        let className = '';

        if (symbol.scope !== null) {
            className = symbol.scope.getName();
        }

        let key = this.getKey(className, symbol.getName());

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key)
        ]);
    }

    async searchByClass(className: string, methodName: string): Promise<Method[]> {
        let methods: Method[] = [];
        let key = this.getKey(className, methodName);

        let uris = await NameIndex.get(this.classIndex, key);

        for (let uri of uris) {
            methods.push(await this.db.get(BelongsToDoc.getKey(uri, key)) as Method);
        }

        return methods;
    }

    async removeByDoc(uri: string) {
        let methods = await BelongsToDoc.removeByDocGetSymbols(this.db, uri) as Method[];

        for (let method of methods) {
            let className = '';

            if (method.scope !== null) {
                className = method.scope.getName();
            }

            await NameIndex.remove(this.classIndex, uri, this.getKey(className, method.getName()));
        }
    }

    private getKey(className: string, methodName: string): string {
        return `${className}${MethodTable.CLASS_SEP}${methodName}`;
    }
}