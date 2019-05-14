import { PhpDocument } from "../../../symbol/phpDocument";
import { Symbol } from "../../../symbol/symbol";
import { DbStore } from "../../db";
import { DbHelper } from "../../dbHelper";

export namespace BelongsToDoc {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string, symbol: Symbol) {
        return db.put(getKey(phpDoc.uri, name), symbol);
    }

    export async function removeByDocGetSymbols(db: DbStore, uri: string): Promise<Symbol[]> {
        let prefix = uri + DbStore.URI_SEP;

        return await DbHelper.deleteInStream<Symbol>(db, db.prefixSearch(prefix), (data) => {
            return data.value;
        });
    }

    export async function removeByDoc(db: DbStore, uri: string): Promise<string[]> {
        let prefix = uri + DbStore.URI_SEP;

        return await DbHelper.deleteInStream<string>(db, db.prefixSearch(prefix), (data) => {
            return data.key.substr(data.key.indexOf(DbStore.URI_SEP));
        });
    }

    export async function get<T>(db: DbStore, uri: string, name: string): Promise<T> {
        return await db.get(getKey(uri, name)) as T;
    }

    export async function getByDoc<T>(db: DbStore, uri: string): Promise<T[]> {
        const prefix = uri + DbStore.URI_SEP;

        return new Promise<T[]>((resolve, reject) => {
            const results: T[] = [];

            db.prefixSearch(prefix)
            .on('data', ({key, value}) => {
                results.push(value);
            })
            .on('error', err => {
                reject();
            })
            .on('end', () => {
                resolve(results);
            });
        });
    }

    export function getKey(uri: string, name: string) {
        return uri + DbStore.URI_SEP + name;
    }
}