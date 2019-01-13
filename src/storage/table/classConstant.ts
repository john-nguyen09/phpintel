import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { ClassConstant } from "../../symbol/constant/classConstant";
import { BelongsToDoc } from "./index/belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./index/nameIndex";
import { CompletionIndex, CompletionValue } from "./index/completionIndex";

@injectable()
export class ClassConstantTable {
    private static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;
    private completionIndex: CompletionIndex;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'classConstant',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.classIndex = new SubStore(level, {
            name: 'classConstantClassIndex',
            version: 1
        });
        this.completionIndex = new CompletionIndex(level, 'classConstantCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: ClassConstant) {
        let className = '';
        if (symbol.scope !== null) {
            className = symbol.scope.name;
        }

        let key = ClassConstantTable.getKey(className, symbol.getName());

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key),
            this.completionIndex.put(phpDoc, symbol.getName(), className)
        ]);
    }

    async getByClass(className: string, constName: string): Promise<ClassConstant[]> {
        let classConsts: ClassConstant[] = [];
        let key = ClassConstantTable.getKey(className, constName);
        let uris = await NameIndex.get(this.classIndex, key);

        for (let uri of uris) {
            classConsts.push(await this.db.get(BelongsToDoc.getKey(uri, key)) as ClassConstant);
        }

        return classConsts;
    }

    async searchAllInClass(className: string): Promise<ClassConstant[]> {
        let classConsts: ClassConstant[] = [];
        let prefix = ClassConstantTable.getKey(className, '');
        let datas = await NameIndex.prefixSearch(this.classIndex, prefix);

        for (let data of datas) {
            const classConst = await this.db.get(BelongsToDoc.getKey(data.uri, data.name)) as ClassConstant;
            classConsts.push(classConst);
        }

        return classConsts;
    }

    async search(className: string, keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword, className);
    }

    async removeByDoc(uri: string) {
        let classConsts = await BelongsToDoc.removeByDocGetSymbols(this.db, uri) as ClassConstant[];

        for (let classConst of classConsts) {
            let className = '';
            if (classConst.scope !== null) {
                className = classConst.scope.name;
            }

            await NameIndex.remove(
                this.classIndex,
                uri,
                ClassConstantTable.getKey(className, classConst.getName())
            );

            await this.completionIndex.del(uri, classConst.getName(), className);
        }
    }

    private static getKey(className: string, constName: string): string {
        return `${className}${this.CLASS_SEP}${constName}`;
    }
}