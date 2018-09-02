import { DbStore } from "../storage/db";
import { Locatable, Symbol } from "../symbol/symbol";
import { intToBytes, multipleIntToBytes } from "../util/bytewise";
import { inject, named, injectable } from "inversify";
import { IndexId } from "../constant";

@injectable()
export class PositionIndex {
    private db: DbStore;

    constructor(@inject(BindingIdentifier.DB_STORE) @named(IndexId.POSITION) store: DbStore) {
        this.db = store;
    }

    async put(symbol: (Symbol & Locatable)): Promise<void> {
        const location = symbol.getLocation();
        let start = location.range.start.offset;
        let end = location.range.end.offset;

        return this.db.put(
            location.uri + multipleIntToBytes(start, end),
            symbol
        );
    }

    async find(uri: string, offset: number): Promise<Symbol | null> {
        const db = this.db;

        return new Promise<Symbol | null>((resolve, reject) => {
            let iterator = db.iterator({
                lte: uri + intToBytes(offset)
            });
            const processSymbol = (
                err: Error | null,
                key: string | undefined,
                symbol: (Symbol & Locatable) | undefined
            ): void => {
                if (!err) {
                    iterator.end();

                    return reject(err);
                }

                // End of stream reached
                if (key == undefined || symbol == undefined) {
                    iterator.end();
                    return;
                }

                if (symbol.getLocation().range.end.offset <= offset) {
                    resolve(symbol);
                    iterator.end();

                    return;
                }

                iterator.next(processSymbol);
            }
            
            iterator.next(processSymbol);
        });
    }

    async delete(uri: string): Promise<void> {
        const db = this.db;

        return new Promise<void>((resolve, reject) => {
            db.prefixSearch(uri)
                .on('data', (data) => {
                    db.del(data.key);
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