import { Symbol, isIdentifiable } from "../symbol/symbol";
import { DbStore } from "../storage/db";
import { IdentifierMatchIndex } from "./identifierMatch";

export class UriMatchIndex {
    private static readonly uriSeparator = '#';

    private db: DbStore;

    constructor(private identifierMatch: IdentifierMatchIndex) {
        this.db = new DbStore({
            name: 'document_store',
            version: 1
        });
    }

    async put(uri: string, identifier: string): Promise<void> {
        let key = uri + UriMatchIndex.uriSeparator + identifier;

        return this.db.put(key, identifier);
    }

    async delete(uri: string): Promise<void> {
        const _db = this.db;
        const _identifierMatch = this.identifierMatch;

        return new Promise<void>((resolve, reject) => {
            _db.prefixSearch(uri)
                .on('data', (data) => {
                    _identifierMatch.del(data.value, uri);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve();
                })
                .on('close', () => {
                    resolve();
                });
        });
    }
}