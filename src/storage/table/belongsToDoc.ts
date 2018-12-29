import { PhpDocument } from "../../symbol/phpDocument";
import { Symbol } from "../../symbol/symbol";
import { DbStore } from "../db";

export namespace BelongsToDoc {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string, symbol: Symbol) {
        return db.put(getKey(phpDoc.uri, name), symbol);
    }

    export async function removeByDocGetSymbols(db: DbStore, uri: string): Promise<Symbol[]> {
        let prefix = uri + DbStore.URI_SEP;

        return new Promise<Symbol[]>((resolve, reject) => {
            let symbols: Symbol[] = [];

            db.prefixSearch(prefix)
                .on('data', (data) => {
                    symbols.push(data.value);

                    db.del(data.key);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve(symbols);
                });
        });
    }

    export async function removeByDoc(db: DbStore, uri: string): Promise<string[]> {
        let prefix = uri + DbStore.URI_SEP;

        return new Promise<string[]>((resolve, reject) => {
            let names: string[] = [];

            db.prefixSearch(prefix)
                .on('data', (data) => {
                    names.push(data.key.substr(data.key.indexOf(DbStore.URI_SEP)));

                    db.del(data.key);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve(names);
                });
        });
    }

    export function getKey(uri: string, name: string) {
        return uri + DbStore.URI_SEP + name;
    }
}