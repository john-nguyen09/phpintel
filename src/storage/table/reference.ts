import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer, Deserializer } from "../serializer";
import { TypeName } from "../../type/name";
import { TypeComposite } from "../../type/composite";
import { injectable } from "inversify";
import { Reference } from "../../symbol/reference";
import { Range } from "../../symbol/meta/range";

@injectable()
export class ReferenceTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'reference',
            version: 1,
            keyEncoding: 'binary',
            valueEncoding: ReferenceEncoding
        });
    }

    async put(ref: Reference) {
        if (ref.location.uri === undefined || ref.location.range === undefined) {
            return;
        }

        let serializer = new Serializer(8);
        serializer.setInt32(ref.location.range.end);
        serializer.setInt32(ref.location.range.start);

        let key = Buffer.concat([
            Buffer.from(ref.location.uri),
            serializer.getBuffer()
        ]);

        return this.db.put(key, ref);
    }

    async removeByDoc(uri: string) {
        const db = this.db;

        return new Promise<void>((resolve, reject) => {
            db.prefixSearch(uri)
                .on('data', (data) => {
                    db.del(data.key);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve();
                });
        });
    }

    async findAt(uri: string, offset: number): Promise<Reference | null> {
        const db = this.db;
        // const logger = App.get<LogWriter>(LogWriter);

        return new Promise<Reference | null>((resolve, reject) => {
            let serializer = new Serializer(4);
            serializer.setInt32(offset);

            let uriBuffer = Buffer.from(uri);

            let key = Buffer.concat([
                uriBuffer,
                serializer.getBuffer()
            ]);
            let iterator = db.iterator<Reference>({
                gte: key,
                lte: Buffer.concat([uriBuffer, Buffer.from('\xFF')])
            });

            const processRef = (
                err: Error | null,
                key?: string | Buffer,
                ref?: Reference
            ): void => {
                if (err) {
                    iterator.end(() => { });

                    return reject(err);
                }

                // End of stream reached
                if (typeof key === 'undefined' || typeof ref === 'undefined') {
                    iterator.end(() => {
                        resolve(null);
                    });
                    return;
                }

                if (
                    ref.location.uri === uri &&
                    ref.location.range !== undefined &&
                    ref.location.range.start <= offset &&
                    ref.location.range.end >= offset
                ) {
                    iterator.end(() => {
                        resolve(ref);
                    });

                    return;
                }

                iterator.next(processRef);
            }

            iterator.next(processRef);
        });
    }

    async findWithin(uri:string, range: Range, predicate?: (ref: Reference) => boolean): Promise<Reference[]> {
        const db = this.db;

        return new Promise<Reference[]>((resolve, reject) => {
            const startSerializer = new Serializer(4);
            startSerializer.setInt32(range.start);
            const endSerializer = new Serializer(4);
            endSerializer.setInt32(range.end);

            const uriBuffer = Buffer.from(uri);
            const iterator = db.iterator<Reference>({
                gte: Buffer.concat([uriBuffer, startSerializer.getBuffer()]),
                lte: Buffer.concat([uriBuffer, endSerializer.getBuffer()]),
            });
            let refs: Reference[] = [];
            const processRef = (err: Error | null, key?: string | Buffer, ref?: Reference): void => {
                if (err !== null) {
                    iterator.end(() => { reject(err); });
                    return;
                }
                if (typeof key === 'undefined' || typeof ref === 'undefined') {
                    iterator.end(() => { resolve(refs); });
                    return;
                }

                if (
                    ref.location.uri === uri &&
                    ref.location.range !== undefined &&
                    ref.location.range.end <= range.end &&
                    ref.scopeRange !== undefined &&
                    ref.scopeRange.start === range.start &&
                    ref.scopeRange.end === range.end &&
                    (typeof predicate === 'undefined' || predicate(ref))
                ) {
                    refs.push(ref);
                }

                iterator.next(processRef);
            }
            iterator.next(processRef);
        });
    }
}

enum TypeKind {
    TYPE_NAME = 1,
    TYPE_COMPOSITE = 2
};

export const ReferenceEncoding = {
    type: 'reference-encoding',
    encode(ref: Reference): Buffer {
        let serializer = new Serializer(128);
        let hasName = ref.refName !== undefined;

        serializer.setBool(hasName);
        if (ref.refName !== undefined) {
            serializer.setString(ref.refName);
        }

        if (ref.type instanceof TypeName) {
            serializer.setInt32(TypeKind.TYPE_NAME);
            serializer.setTypeName(ref.type);
        } else if (ref.type instanceof TypeComposite) {
            serializer.setInt32(TypeKind.TYPE_COMPOSITE);
            serializer.setTypeComposite(ref.type);
        }

        serializer.setLocation(ref.location);
        serializer.setInt32(ref.refKind);

        const hasScope = ref.scope !== null;

        serializer.setBool(hasScope);
        if (hasScope) {
            if(ref.scope instanceof TypeName) {
                serializer.setInt32(TypeKind.TYPE_NAME);
                serializer.setTypeName(ref.scope);
            } else if (ref.scope instanceof TypeComposite) {
                serializer.setInt32(TypeKind.TYPE_COMPOSITE);
                serializer.setTypeComposite(ref.scope);
            }
        }

        if (ref.scopeRange === undefined) {
            serializer.setBool(false);
        } else {
            serializer.setBool(true);
            serializer.setRange(ref.scopeRange);
        }

        return serializer.getBuffer();
    },
    decode(buffer: Buffer): Reference | null {
        if (buffer.length === 0) {
            return null;
        }

        const deserializer = new Deserializer(buffer);
        let type: TypeName | TypeComposite = new TypeName('');
        let scope: TypeName | TypeComposite | null = null;
        const hasName = deserializer.readBool();
        let refName: string | undefined = undefined;

        if (hasName) {
            refName = deserializer.readString();
        }

        const typeKind: TypeKind = deserializer.readInt32();

        if (typeKind === TypeKind.TYPE_NAME) {
            type = deserializer.readTypeName() || new TypeName('');
        } else if (typeKind === TypeKind.TYPE_COMPOSITE) {
            type = deserializer.readTypeComposite();
        }

        const location = deserializer.readLocation();
        const refKind = deserializer.readInt32();

        const hasScope = deserializer.readBool();
        if (hasScope) {
            const typeKind = deserializer.readInt32();
            if (typeKind === TypeKind.TYPE_NAME) {
                scope = deserializer.readTypeName();
            } else if (typeKind === TypeKind.TYPE_COMPOSITE) {
                scope = deserializer.readTypeComposite();
            }
        }

        const hasScopeRange = deserializer.readBool();
        let scopeRange: Range | undefined = undefined;

        if (hasScopeRange) {
            scopeRange = deserializer.readRange();
        }

        return {
            refName,
            type,
            location,
            refKind,
            scope,
            scopeRange
        } as Reference;
    },
    buffer: true
};