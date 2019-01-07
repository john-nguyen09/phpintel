import { DbStore, LevelDatasource, SubStore } from "../db";
import { Serializer } from "../serializer";
import { injectable } from "inversify";
import { PhpDocument } from "../../symbol/phpDocument";

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
}

export const PhpDocEncoding = {
    type: 'php-doc-encoding',
    encode(phpDoc: PhpDocument): Buffer {
        let serializer = new Serializer;

        serializer.writeString(phpDoc.uri);
        serializer.writeString(phpDoc.text);
        serializer.writeInt32(phpDoc.modifiedTime);

        return serializer.getBuffer();
    },
    decode(buffer: Buffer): PhpDocument {
        let serializer = new Serializer(buffer);
        let phpDoc = new PhpDocument(serializer.readString(), serializer.readString());
        phpDoc.modifiedTime = serializer.readInt32();

        return phpDoc;
    },
    buffer: true
} as Level.Encoding;