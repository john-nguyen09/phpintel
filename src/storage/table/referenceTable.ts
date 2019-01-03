import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer } from "../serializer";
import { TypeName } from "../../type/name";
import { TypeComposite } from "../../type/composite";
import { injectable } from "inversify";
import { Reference } from "../../symbol/reference";

@injectable()
export class ReferenceTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'reference',
            version: 1,
            valueEncoding: ReferenceEncoding
        });
    }

    async put(reference: Reference) {
        if (reference.location.isEmpty) {
            return;
        }

        let serializer = new Serializer();
        serializer.writeInt32(reference.location.range.start.offset);
        serializer.writeInt32(reference.location.range.end.offset);

        let key = Buffer.concat([
            Buffer.from(reference.location.uri),
            serializer.getBuffer()
        ]);

        return this.db.put(key, reference);
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
            let serializer = new Serializer();
            serializer.writeInt32(offset);

            let uriBuffer = Buffer.from(uri);

            let key = Buffer.concat([
                uriBuffer,
                serializer.getBuffer(),
                Buffer.from('\xFF')
            ]);
            let iterator = db.iterator<Reference>({
                lte: key,
                gte: uriBuffer
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
                if (key == undefined || ref == undefined) {
                    iterator.end(() => {
                        resolve(null);
                    });
                    return;
                }

                // logger.info(JSON.stringify(refObject(ref)));
                // logger.info(ref.location.range.end.offset.toString());
                // logger.info(offset.toString());
                // logger.info(JSON.stringify(ref.location.range.end.offset >= offset));

                if (ref.location.range.end.offset >= offset) {
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
}

enum TypeKind {
    TYPE_NAME = 1,
    TYPE_COMPOSITE = 2
};

export const ReferenceEncoding = {
    type: 'reference-encoding',
    encode(ref: Reference): Buffer {
        let serializer = new Serializer();
        let hasName = ref.refName !== undefined;

        serializer.writeBool(hasName);
        if (ref.refName !== undefined) {
            serializer.writeString(ref.refName);
        }

        if (ref.type instanceof TypeName) {
            serializer.writeInt32(TypeKind.TYPE_NAME);
            serializer.writeTypeName(ref.type);
        } else if (ref.type instanceof TypeComposite) {
            serializer.writeInt32(TypeKind.TYPE_COMPOSITE);
            serializer.writeTypeComposite(ref.type);
        }

        serializer.writeLocation(ref.location);
        serializer.writeInt32(ref.refKind);
        serializer.writeTypeName(ref.scope);

        return serializer.getBuffer();
    },
    decode(buffer: Buffer): Reference | null {
        if (buffer.length === 0) {
            return null;
        }

        let serializer = new Serializer(buffer);
        let type: TypeName | TypeComposite = new TypeName('');
        let hasName = serializer.readBool();
        let refName: string | undefined = undefined;

        if (hasName) {
            refName = serializer.readString();
        }

        let typeKind: TypeKind = serializer.readInt32();

        if (typeKind == TypeKind.TYPE_NAME) {
            type = serializer.readTypeName() || new TypeName('');
        } else if (typeKind === TypeKind.TYPE_COMPOSITE) {
            type = serializer.readTypeComposite();
        }

        let location = serializer.readLocation();
        let refKind = serializer.readInt32();
        let scope = serializer.readTypeName();
        
        return {
            refName,
            type,
            location,
            refKind,
            scope
        } as Reference;
    }
} as Level.Encoding;