import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer } from "../serializer";
import { injectable } from "inversify";
import { PhpDocument } from "../../symbol/phpDocument";
import { ImportTable } from "../../type/importTable";

@injectable()
export class PhpDocumentTable {
    private db: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'php_document',
            version: 1,
            valueEncoding: PhpDocEncoding
        });
    }

    async put(phpDoc: PhpDocument) {
        return this.db.put(phpDoc.uri, phpDoc);
    }

    async get(uri: string): Promise<PhpDocument | null> {
        try {
            return await this.db.get(uri);
        } catch {
            return null;
        }
    }

    async remove(uri: string) {
        return this.db.del(uri);
    }

    async getAllStream<T>(callback: (phpDoc: PhpDocument) => void) {
        const db = this.db;

        return await new Promise<void>((resolve, reject) => {
            db.createReadStream()
                .on('data', (data) => {
                    callback(data.value);
                })
                .on('end', () => {
                    resolve();
                })
                .on('error', (err) => {
                    if (err) {
                        reject(err);
                    }
                });
        });
    }
}

export const PhpDocEncoding = {
    type: 'php-doc-encoding',
    encode(phpDoc: PhpDocument): Buffer {
        let serializer = new Serializer;

        serializer.writeString(phpDoc.uri);
        serializer.writeString(phpDoc.text);
        serializer.writeInt32(phpDoc.modifiedTime);
        
        serializer.writeNamespaceName(phpDoc.importTable.namespace);

        let keys = Object.keys(phpDoc.importTable.imports);
        serializer.writeInt32(keys.length);
        for (let key of keys) {
            serializer.writeString(key);
            serializer.writeString(phpDoc.importTable.imports[key]);
        }

        return serializer.getBuffer();
    },
    decode(buffer: Buffer): PhpDocument {
        let serializer = new Serializer(buffer);
        let phpDoc = new PhpDocument(serializer.readString(), serializer.readString());
        phpDoc.modifiedTime = serializer.readInt32();

        let importTable = new ImportTable();
        importTable.namespace = serializer.readNamespaceName();
        let numOfKeys = serializer.readInt32();

        for (let i = 0; i < numOfKeys; i++) {
            importTable.imports[serializer.readString()] = serializer.readString();
        }

        phpDoc.importTable = importTable;

        return phpDoc;
    },
    buffer: true
} as Level.Encoding;