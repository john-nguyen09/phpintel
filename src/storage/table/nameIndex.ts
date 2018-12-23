import { DbStore } from "../db";
import { PhpDocument } from "../../symbol/phpDocument";

export namespace NameIndex {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string) {
        return db.put(name + DbStore.uriSep + phpDoc.uri, phpDoc.uri + DbStore.uriSep + name);
    }

    export async function remove(db: DbStore, uri: string, name: string) {
        return db.del(name + DbStore.uriSep + uri);
    }

    export async function get(db: DbStore, name: string): Promise<string[]> {
        return new Promise<string[]>((resolve, reject) => {
            let keys: string[] = [];

            db.prefixSearch(name + DbStore.uriSep)
                .on('data', (data) => {
                    keys.push(data.value);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve(keys);
                });
        });
    }
}