import { DbStore, LevelDatasource, SubStore } from "../db";
import { injectable } from "inversify";
import { ArgumentExpressionList } from "../../symbol/argumentExpressionList";
import * as bytewise from "bytewise";
import { PhpDocument } from "../../symbol/phpDocument";
import { Range } from "../../symbol/meta/range";
import { DbHelper } from "../dbHelper";
import { TypeComposite } from "../../type/composite";
import { TypeKind } from "./reference";
import { TypeName } from "../../type/name";

@injectable()
export class ArgumentListTable {
    private db: DbStore;
    private belongsToDoc: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'argument_list',
            version: 1,
            keyEncoding: bytewise,
            valueEncoding: ArgumentListEncoding,
        });
    }

    async put(phpDoc: PhpDocument, argumentList: ArgumentExpressionList) {
        if (argumentList.location.range === undefined) {
            return;
        }

        return await this.db.put([
            phpDoc.uri,
            argumentList.location.range.end,
            argumentList.location.range.start,
        ], argumentList);
    }

    async removeByDoc(uri: string) {
        return await DbHelper.deleteInStream<void>(this.db, this.db.createReadStream({
            gte: [uri],
            lte: [uri, '\xFF'],
        }));
    }

    async findAt(uri: string, offset: number): Promise<ArgumentExpressionList | null> {
        return new Promise<ArgumentExpressionList | null>((resolve, reject) => {
            const iterator = this.db.iterator<ArgumentExpressionList>({
                gte: [uri, offset],
                lte: [uri, '\xFF']
            });

            const processArgumentList = (
                err: Error | null,
                key?: any,
                argumentList?: ArgumentExpressionList
            ): void => {
                if (err) {
                    iterator.end(() => { reject(err); });
                    return;
                }

                if (key === undefined || argumentList === undefined) {
                    iterator.end(() => { resolve(null) });
                    return;
                }

                if (
                    argumentList.location.uri === uri &&
                    argumentList.location.range !== undefined &&
                    argumentList.location.range.start <= offset &&
                    argumentList.location.range.end >= offset
                ) {
                    iterator.end(() => { resolve(argumentList) });
                    return;
                }

                iterator.next(processArgumentList);
            }

            iterator.next(processArgumentList);
        });
    }
}

const ArgumentListEncoding: Level.Encoding = {
    type: 'argument_list',
    encode: (argumentList: ArgumentExpressionList): any => {
        const array: (number | string | boolean)[] = [];

        if (argumentList.location.uri === undefined) {
            array.push(false);
        } else {
            array.push(true);
            array.push(argumentList.location.uri);
        }

        if (argumentList.location.range === undefined) {
            array.push(false);
        } else {
            array.push(true);
            array.push(argumentList.location.range.start);
            array.push(argumentList.location.range.end);
        }

        array.push(argumentList.commaOffsets.length);
        for (const commaOffset of argumentList.commaOffsets) {
            array.push(commaOffset);
        }

        array.push(argumentList.type.name);

        if (argumentList.scope !== null) {
            array.push(true);
            if (argumentList.scope instanceof TypeComposite) {
                array.push(TypeKind.TYPE_COMPOSITE);
                array.push(argumentList.scope.types.length);
                for (const type of argumentList.scope.types) {
                    array.push(type.name);
                }
            } else {
                array.push(TypeKind.TYPE_NAME);
                array.push(argumentList.scope.name);
            }
        } else {
            array.push(false);
        }

        return bytewise.encode(array);
    },
    decode: (buffer: any): ArgumentExpressionList => {
        const array: (number | string | boolean)[] = bytewise.decode(buffer);
        const argumentList = new ArgumentExpressionList();

        if (array.shift()) {
            argumentList.location.uri = (array.shift() as string);
        }

        if (array.shift()) {
            argumentList.location.range = {
                start: array.shift() as number,
                end: array.shift() as number,
            };
        }

        const commaOffsetsLength = array.shift() as number;
        for (let i = 0; i < commaOffsetsLength; i++) {
            argumentList.commaOffsets.push(array.shift() as number);
        }

        let typeName = array.shift();
        if (typeof typeName !== 'string') {
            typeName = '';
        }
        argumentList.type = new TypeName(typeName);

        if (array.shift()) { // has scope
            const typeKind = array.shift() as TypeKind;

            if (typeKind === TypeKind.TYPE_COMPOSITE) {
                const typesLength = array.shift() as number;
                argumentList.scope = new TypeComposite();
                for (let i = 0; i < typesLength; i++) {
                    argumentList.scope.types.push(new TypeName(array.shift() as string));
                }
            } else {
                argumentList.scope = new TypeName(array.shift() as string);
            }
        }

        return argumentList;
    },
    buffer: true
};