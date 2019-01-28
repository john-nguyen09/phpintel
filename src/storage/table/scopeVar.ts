import { injectable } from "inversify";
import { LevelDatasource, DbStore, SubStore } from "../db";
import { ScopeVar } from "../../symbol/variable/scopeVar";
import { Serializer, Deserializer } from "../serializer";
import { Range } from "../../symbol/meta/range";

@injectable()
export class ScopeVarTable {
    public static readonly URI_SEP = '#';

    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'scopeVar',
            version: 1,
            keyEncoding: 'binary',
            valueEncoding: ValueEncoding,
        });
    }

    async put(scopeVar: ScopeVar) {
        if (
            scopeVar.location.range === undefined ||
            scopeVar.location.uri === undefined
        ) {
            return;
        }

        const serializer = new Serializer();
        serializer.setInt32(scopeVar.location.range.end);
        serializer.setInt32(scopeVar.location.range.start);
        const key = Buffer.concat([
            Buffer.from(scopeVar.location.uri),
            serializer.getBuffer()
        ]);

        return await this.db.put(key, scopeVar);
    }

    async removeByDoc(uri: string) {
        const db = this.db;

        return new Promise<void>((resolve, reject) => {
            db.prefixSearch(uri)
                .on('data', (data) => {
                    db.del(data.key);
                })
                .on('error', (err) => {
                    if (err) {
                        reject(err);
                    }
                }).on('end', () => {
                    resolve();
                });
        });
    }

    async findAt(uri: string, offset: number): Promise<Range | null> {
        const db = this.db;

        return new Promise<Range | null>((resolve, reject) => {
            const serializer = new Serializer(4);
            serializer.setInt32(offset);
            const uriBuffer = Buffer.from(uri);
            const key = Buffer.concat([
                uriBuffer,
                serializer.getBuffer()
            ]);

            const iterator = db.iterator<ScopeVar>({
                gte: key,
                lte: Buffer.concat([uriBuffer, Buffer.from('\xFF')]),
            });
            const processScopeVar = (
                err: Error | null,
                key?: string | Buffer,
                scopeVar?: ScopeVar
            ) => {
                if (err !== null) {
                    iterator.end(() => {});
                    return reject(err);
                }

                if (typeof key === 'undefined' || typeof scopeVar === 'undefined') {
                    iterator.end(() => { resolve(null) });
                    return;
                }

                if (
                    scopeVar.location.range !== undefined &&
                    scopeVar.location.range.start <= offset &&
                    scopeVar.location.range.end >= offset
                ) {
                    iterator.end(() => {
                        resolve(scopeVar.location.range);
                    });
                    return;
                }

                iterator.next(processScopeVar);
            }
            iterator.next(processScopeVar);
        });
    }
}

const ValueEncoding = {
    encode: (scopeVar: ScopeVar): Buffer => {
        const serializer = new Serializer();

        if (scopeVar.location.uri === undefined || scopeVar.location.range === undefined) {
            serializer.setBool(false);
        } else {
            serializer.setBool(true);
            serializer.setString(scopeVar.location.uri);
            serializer.setInt32(scopeVar.location.range.start);
            serializer.setInt32(scopeVar.location.range.end);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): ScopeVar => {
        const scopeVar = new ScopeVar();
        const deserializer = new Deserializer(buffer);

        if (deserializer.readBool()) {
            scopeVar.location.uri = deserializer.readString();
            scopeVar.location.range =  {
                start: deserializer.readInt32(),
                end: deserializer.readInt32(),
            }
        }

        return scopeVar;
    }
} as Level.Encoding;