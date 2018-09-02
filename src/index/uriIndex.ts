import { DbStore } from "../storage/db";
import { inject, named, injectable } from "inversify";
import { IndexId } from "../constant";
import { IdentifierIndex } from "./identifierIndex";

@injectable()
export class UriIndex {
    private static readonly uriSeparator = '#';

    private db: DbStore;
    private identifierIndex: IdentifierIndex;

    constructor(
        @inject(BindingIdentifier.DB_STORE) @named(IndexId.URI) store: DbStore,
        @inject(BindingIdentifier.IDENTIFIER_INDEX) identifierIndex: IdentifierIndex
    ) {
        this.db = store;
        this.identifierIndex = identifierIndex;
    }

    async put(uri: string, identifier: string): Promise<void> {
        let key = uri + UriIndex.uriSeparator + identifier;

        return this.db.put(key, identifier);
    }

    async delete(uri: string): Promise<void> {
        const _db = this.db;
        const _identifierMatch = this.identifierIndex;

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