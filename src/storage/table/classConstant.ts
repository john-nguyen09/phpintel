import { DbStore, LevelDatasource, SubStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";
import { ClassConstant } from "../../symbol/constant/classConstant";
import { BelongsToDoc } from "./belongsToDoc";
import { injectable } from "inversify";
import { NameIndex } from "./nameIndex";

@injectable()
export class ClassConstantTable {
    private static readonly CLASS_SEP = '@';

    private db: DbStore;
    private classIndex: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'class_constant',
            version: 1,
            valueEncoding: require('../symbolEncoding')
        });
        this.classIndex = new SubStore(level, {
            name: 'class_constant_class_index',
            version: 1
        });
    }

    async put(phpDoc: PhpDocument, symbol: ClassConstant) {
        let className = '';
        if (symbol.scope !== null) {
            className = symbol.scope.name;
        }

        let key = ClassConstantTable.getKey(className, symbol.getName());

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, key, symbol),
            NameIndex.put(this.classIndex, phpDoc, key)
        ]);
    }

    async searchByClass(className: string, constName: string): Promise<ClassConstant[]> {
        let classConsts: ClassConstant[] = [];
        let key = ClassConstantTable.getKey(className, constName);
        let uris = await NameIndex.get(this.classIndex, key);

        for (let uri of uris) {
            classConsts.push(await this.db.get(BelongsToDoc.getKey(uri, key)) as ClassConstant);
        }

        return classConsts;
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
        }
    }

    private static getKey(className: string, constName: string): string {
        return `${className}${this.CLASS_SEP}${constName}`;
    }
}