import { Symbol, isLocatable, isIdentifiable, Identifiable } from "../symbol/symbol";
import { DbStore } from "../storage/db";
import { IdentifierMatchIndex } from "./identifierMatch";
import { multipleIntToBytes, intToBytes } from "../util/bytewise";
import { Location } from "../symbol/meta/location";

export class LocationMatchIndex {
    private static readonly uriGlue = '#';

    private db: DbStore;

    constructor(private identifierMatch: IdentifierMatchIndex) {
        this.db = new DbStore({
            name: 'location_store',
            version: 1
        });
    }

    async put(symbol: (Symbol & Identifiable)): Promise<void> {
        if (!isLocatable(symbol)) {
            return;
        }

        let location = symbol.getLocation();
        let key = location.uri +
            LocationMatchIndex.uriGlue +
            multipleIntToBytes(location.range.start.offset, location.range.end.offset);

        return this.db.put(key, symbol.getIdentifier());
    }

    async get(uri: string, start: number, end: number): Promise<Symbol[]> {
        const _db = this.db;

        let identifiers = await new Promise<string[]>((resolve, reject) => {
            let identifiers: string[] = [];

            _db.createReadableStream({
                gte: uri + LocationMatchIndex.uriGlue + intToBytes(start),
                lte: uri + LocationMatchIndex.uriGlue + intToBytes(end)
            }).on('data', (data: {key: string, value: string}) => {
                identifiers.push(data.value);
            }).on('end', () => {
                resolve(identifiers);
            }).on('error', (err) => {
                reject(err);
            });
        });

        let results: Symbol[] = [];

        for (let identifier of identifiers) {
            let symbols = await this.identifierMatch.get(identifier);

            for (let symbol of symbols) {
                if (!isLocatable(symbol)) {
                    continue;
                }

                if (symbol.getLocation().range.end.offset > end) {
                    continue;
                }

                results.push(symbol);
            }
        }

        return results;
    }

    async delete(uri: string): Promise<void> {
        const _db = this.db;

        return new Promise<void>((resolve, reject) => {
            _db.prefixSearch(uri + LocationMatchIndex.uriGlue)
                .on('data', (data: {key: string, value: string}) => {
                    _db.del(data.key);
                }).on('end', () => {
                    resolve();
                }).on('error', (err) => {
                    reject(err);
                });
        });
    }
}