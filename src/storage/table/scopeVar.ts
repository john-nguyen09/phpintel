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
        const serializer = new Serializer();
        serializer.setInt32(scopeVar.location.range.start);
        serializer.setInt32(scopeVar.location.range.end);
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

    async findBefore(uri: string, offset: number): Promise<ScopeVar | null> {
        const db = this.db;

        return new Promise<ScopeVar | null>((resolve, reject) => {
            const serializer = new Serializer();
            serializer.setInt32(offset);
            const uriBuffer = Buffer.from(uri);
            const key = Buffer.concat([
                uriBuffer,
                serializer.getBuffer(),
                Buffer.from('\xFF')
            ]);

            const iterator = db.iterator<ScopeVar>({
                lte: key,
                gte: uriBuffer,
                reverse: true
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

                iterator.end(() => {
                    resolve(scopeVar);
                });
            }
            iterator.next(processScopeVar);
        });
    }

    async findAfter(uri: string, offset: number): Promise<ScopeVar | null> {
        const db = this.db;

        return new Promise<ScopeVar | null>((resolve, reject) => {
            const serializer = new Serializer();
            serializer.setInt32(offset);
            const uriBuffer = Buffer.from(uri);

            const iterator = db.iterator<ScopeVar>({
                gte: Buffer.concat([uriBuffer, serializer.getBuffer()]),
                lte: Buffer.concat([uriBuffer, Buffer.from('\xFF')])
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

                iterator.end(() => {
                    resolve(scopeVar);
                });
            }
            iterator.next(processScopeVar);
        });
    }

    async findAt(uri: string, offset: number): Promise<Range> {
        const results = await Promise.all([
            this.findBefore(uri, offset),
            this.findAfter(uri, offset)
        ]);
        const before = results[0];
        const after = results[1];

        if (before === null) {
            throw new Error(`[Impossble case]: Cannot find any scope before ${offset}`);
        }
        if (after === null) {
            return before.location.range;
        }

        return {
            start: before.location.range.start,
            end: after.location.range.start
        };
    }
}

const ValueEncoding = {
    encode: (scopeVar: ScopeVar): Buffer => {
        const serializer = new Serializer();

        serializer.setString(scopeVar.location.uri);
        if (typeof scopeVar.location.range !== 'undefined') {
            serializer.setBool(true);
            serializer.setInt32(scopeVar.location.range.start);
            serializer.setInt32(scopeVar.location.range.end);
        } else {
            serializer.setBool(false);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): ScopeVar => {
        const scopeVar = new ScopeVar();
        const deserializer = new Deserializer(buffer);

        scopeVar.location.uri = deserializer.readString();
        if (deserializer.readBool()) {
            scopeVar.location.range =  {
                start: deserializer.readInt32(),
                end: deserializer.readInt32(),
            }
        }

        return scopeVar;
    }
} as Level.Encoding;