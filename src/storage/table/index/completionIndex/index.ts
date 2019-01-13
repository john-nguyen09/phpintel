import { WordSeparator } from "./wordSeparator";
import { DbStore, LevelDatasource, SubStore } from "../../../db";
import { PhpDocument } from "../../../../symbol/phpDocument";
import { Serializer } from "../../../serializer";

export interface CompletionValue {
    uri: string;
    name: string;
}

export class CompletionIndex {
    public static readonly INFO_SEP = '#';
    public static LIMIT = 100;

    private db: DbStore;

    constructor(datasource: LevelDatasource, name: string) {
        this.db = new SubStore(datasource, {
            name: name,
            version: 1,
            keyEncoding: 'binary',
            valueEncoding: CompletionEncoding
        });
    }

    async put(phpDoc: PhpDocument, name: string, prefix?: string) {
        let tokens = WordSeparator.getTokens(name);

        for (let token of tokens) {
            let indexKey = CompletionIndex.getKey(phpDoc.uri, token);
            if (typeof prefix !== 'undefined') {
                indexKey = prefix + indexKey;
            }

            await this.db.put(indexKey, {
                uri: phpDoc.uri,
                name: name
            });
        }
    }

    async search(keyword: string, prefix?: string): Promise<CompletionValue[]> {
        const db = this.db;
        let completions: CompletionValue[] = [];

        if (typeof prefix !== 'undefined') {
            keyword = prefix + keyword;
        }

        return new Promise<CompletionValue[]>((resolve, reject) => {
            let readStream: NodeJS.ReadableStream;

            if (keyword.length === 0) {
                readStream = db.createReadStream({
                    limit: CompletionIndex.LIMIT
                });
            } else {
                readStream = db.prefixSearch(keyword, CompletionIndex.LIMIT);
            }

            readStream
                .on('data', (data) => {
                    completions.push(data.value);
                })
                .on('end', () => {
                    resolve(completions);
                })
                .on('reject', (err) => {
                    if (err) {
                        reject(err);
                    }
                });
        });
    }

    async del(uri: string, name: string, prefix?: string) {
        let tokens = WordSeparator.getTokens(name);

        for (let token of tokens) {
            let key = CompletionIndex.getKey(uri, token);
            await this.db.del(CompletionIndex.getKey(uri, token));
        }
    }

    public static getKey(uri: string, token: string) {
        return `${token}${CompletionIndex.INFO_SEP}${uri}`;
    }
}

const CompletionEncoding = {
    type: 'completion-encoding',
    encode: (value: CompletionValue): Buffer => {
        let serializer = new Serializer();

        serializer.writeString(value.uri);
        serializer.writeString(value.name);

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): CompletionValue => {
        let serializer = new Serializer(buffer);

        return {
            uri: serializer.readString(),
            name: serializer.readString()
        };
    }
} as Level.Encoding;