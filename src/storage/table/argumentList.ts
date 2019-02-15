import { DbStore, LevelDatasource, SubStore } from "../db";
import { injectable } from "inversify";
import { ArgumentExpressionList } from "../../symbol/argumentExpressionList";
import * as bytewise from "bytewise";
import { PhpDocument } from "../../symbol/phpDocument";
import { Range } from "../../symbol/meta/range";
import { DbHelper } from "../dbHelper";

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

        for (const range of argumentList.ranges) {
            array.push(range.start);
            array.push(range.end);
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

        if ((array.length % 2) !== 0) {
            throw Error('ArgumentListEncoding received invalid buffer');
        }

        for (let i = 0; i < array.length; i += 2) {
            argumentList.ranges.push({
                start: array[i] as number,
                end: array[i + 1] as number,
            });
        }

        return argumentList;
    },
    buffer: true
};