import { WordSeparator } from "./wordSeparator";
import { DbStore, LevelDatasource, SubStore } from "../../../db";
import { PhpDocument } from "../../../../symbol/phpDocument";
import { Serializer, Deserializer } from "../../../serializer";
import { inspect } from "util";

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
            valueEncoding: CompletionEncoding
        });
    }

    async put(phpDoc: PhpDocument, name: string, prefix?: string) {
        if (typeof name !== 'string') {
            console.trace(`${phpDoc.uri} invalid put, name is not string ${inspect(name)}`);
            return;
        }

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
        if (typeof name !== 'string') {
            return;
        }

        let tokens = WordSeparator.getTokens(name);

        if (typeof prefix === 'undefined') {
            prefix = '';
        }

        for (let token of tokens) {
            await this.db.del(prefix + CompletionIndex.getKey(uri, token));
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

        serializer.setString(value.uri);
        serializer.setString(value.name);

        return serializer.getBuffer();
    },
    decode: (buffer: Buffer): CompletionValue => {
        let deserializer = new Deserializer(buffer);

        return {
            uri: deserializer.readString(),
            name: deserializer.readString()
        };
    }
} as Level.Encoding;