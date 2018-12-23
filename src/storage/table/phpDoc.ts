import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer } from "../serializer";
import { injectable } from "inversify";
import { App } from "../../app";
import { LogWriter } from "../../service/logWriter";

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

    async put(uri: string, lastModified: number) {
        return this.db.put(uri, lastModified);
    }

    async get(uri: string): Promise<number> {
        try {
            let lastUpdated = await this.db.get(uri);

            return lastUpdated;
        } catch (err) {
            return -1;
        }
    }
}

const PhpDocEncoding = {
    type: 'php-doc-encoding',
    encode(lastModified: number): Buffer {
        let serializer = new Serializer;

        serializer.writeInt32(lastModified);

        return serializer.getBuffer();
    },
    decode(buffer: Buffer): number {
        let serializer = new Serializer(buffer);

        return serializer.readInt32();
    },
    buffer: true
} as Level.Encoding;