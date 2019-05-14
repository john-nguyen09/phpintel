import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer, Deserializer } from "../serializer";
import { injectable } from "inversify";
import { PhpDocument } from "../../symbol/phpDocument";
import { ImportTable } from "../../type/importTable";
import * as AsyncLock from "async-lock";

@injectable()
export class PhpDocumentTable {
    private static lock = new AsyncLock();

    private db: DbStore;
    private openedDocs: Map<string, PhpDocument> = new Map<string, PhpDocument>();

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'php_document',
            version: 1,
            valueEncoding: PhpDocEncoding
        });
    }

    async put(phpDoc: PhpDocument, isOpen: boolean) {
        if (isOpen) {
            this.openedDocs.set(phpDoc.uri, phpDoc);
            return;
        }

        return this.db.put(phpDoc.uri, phpDoc);
    }

    async get(uri: string): Promise<PhpDocument | null> {
        try {
            const phpDoc = this.openedDocs.get(uri);

            if (typeof phpDoc !== 'undefined') {
                return phpDoc;
            }

            return await this.db.get(uri);
        } catch {
            return null;
        }
    }

    async remove(uri: string): Promise<void> {
        if (this.openedDocs.has(uri)) {
            this.openedDocs.delete(uri);
            return;
        }

        return this.db.del(uri);
    }

    async getAllStream<T>(callback: (phpDoc: PhpDocument) => void): Promise<void> {
        const db = this.db;

        for (const uri in this.openedDocs) {
            callback(this.openedDocs[uri]);
        }

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

    static async acquireLock(uri: string, action: () => void | PromiseLike<void>): Promise<void> {
        return PhpDocumentTable.lock.acquire(uri, action);
    }
}

export const PhpDocEncoding: Level.Encoding = {
    type: 'php-doc-encoding',
    encode(phpDoc: PhpDocument): string {
        let serializer = new Serializer();

        serializer.setString(phpDoc.uri);
        serializer.setString(phpDoc.text);
        serializer.setInt32(phpDoc.modifiedTime);

        serializer.setNamespaceName(phpDoc.importTable.namespace);

        let keys = Object.keys(phpDoc.importTable.imports);
        serializer.setInt32(keys.length);
        for (let key of keys) {
            serializer.setString(key);
            serializer.setString(phpDoc.importTable.imports[key]);
        }

        return serializer.getBuffer();
    },
    decode(buffer: string): PhpDocument {
        let deserializer = new Deserializer(buffer);
        let phpDoc = new PhpDocument(deserializer.readString(), deserializer.readString());
        phpDoc.modifiedTime = deserializer.readInt32();

        let importTable = new ImportTable();
        importTable.namespace = deserializer.readNamespaceName();
        let numOfKeys = deserializer.readInt32();

        for (let i = 0; i < numOfKeys; i++) {
            importTable.imports[deserializer.readString()] = deserializer.readString();
        }

        phpDoc.importTable = importTable;

        return phpDoc;
    },
    buffer: false
};