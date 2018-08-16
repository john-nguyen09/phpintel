import { DbStore } from "../storage/db";
import { Symbol, isIdentifiable, Identifiable } from "../symbol/symbol";

export class IdentifierMatchIndex {
    private static readonly nameGlue = '#';

    private db: DbStore;

    constructor() {
        this.db = new DbStore({
            name: 'symbol_store',
            version: 1
        });
    }

    async put(symbol: (Symbol & Identifiable), uri: string): Promise<void> {
        return this.db.put(this.getKey(symbol.getIdentifier(), uri), symbol);
    }

    async get(identifier: string): Promise<Symbol[]> {
        const _db = this.db;

        return new Promise<Symbol[]>((resolve, reject) => {
            let symbols: Symbol[] = [];

            _db.prefixSearch(this.getKey(identifier))
                .on('data', (data: {key: string, value: string}) => {
                    symbols.push(JSON.parse(data.value) as Symbol);
                }).on('end', () => {
                    resolve(symbols);
                }).on('error', (err) => {
                    reject(err);
                });
        });
    }

    async del(identifier: string, uri: string): Promise<void> {
        return this.db.del(this.getKey(identifier, uri));
    }

    private getKey(identifier: string, uri?: string) {
        if (uri == undefined) {
            uri = '';
        }

        return identifier + IdentifierMatchIndex.nameGlue + uri;
    }
}