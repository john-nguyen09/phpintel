import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { Method } from "../../symbol/function/method";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";
import { CompletionIndex, CompletionValue } from "./index/completionIndex";
import { Indexer } from "../../index/indexer";
import { App } from "../../app";

export type MethodPredicate = (method: Method) => boolean;

@injectable()
export class MethodTable {
    public static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;
    private completionIndex: CompletionIndex;

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
        this.completionIndex = new CompletionIndex(level, 'methodCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: Method) {
        let className = '';

        if (symbol.scope !== null) {
            className = symbol.scope.name;
        }

        let key = this.getKey(className, symbol.getName());

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key),
            this.completionIndex.put(phpDoc, symbol.getName(), className)
        ]);
    }

    async getByClass(className: string, methodName: string, predicate?: MethodPredicate): Promise<Method[]> {
        let key = this.getKey(className, methodName);

        const uris = await NameIndex.get(this.classIndex, key);

        let methods = await BelongsToDoc.getMultiple<Method>(this.db, uris, key);

        if (predicate !== undefined) {
            methods = methods.filter(predicate);
        }

        return methods;
    }

    async searchAllInClass(className: string, predicate?: MethodPredicate): Promise<Method[]> {
        let methods: Method[] = [];
        const prefix = this.getKey(className, '');
        const datas = await NameIndex.prefixSearch(this.classIndex, prefix);

        methods = await BelongsToDoc.getMultipleByNameIndex<Method>(this.db, datas);

        if (predicate !== undefined) {
            methods = methods.filter(predicate);
        }

        return methods;
    }

    async search(className: string, keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword, className);
    }

    async getByDoc(phpDoc: PhpDocument): Promise<Method[]> {
        const indexer: Indexer = App.get<Indexer>(Indexer);

        if (indexer.isOpen(phpDoc.uri)) {
            return phpDoc.methods;
        }

        return BelongsToDoc.getByDoc<Method>(this.db, phpDoc.uri);
    }

    async removeByDoc(uri: string) {
        let methods = await BelongsToDoc.removeByDocGetSymbols(this.db, uri) as Method[];

        for (let method of methods) {
            let className = '';

            if (method.scope !== null) {
                className = method.scope.name;
            }

            await NameIndex.remove(this.classIndex, uri, this.getKey(className, method.getName()));
            await this.completionIndex.del(uri, method.getName(), className);
        }
    }

    private getKey(className: string, methodName: string): string {
        return `${className}${MethodTable.CLASS_SEP}${methodName}`;
    }
}