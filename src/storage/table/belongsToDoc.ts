import { PhpDocument } from "../../symbol/phpDocument";
import { Symbol } from "../../symbol/symbol";
import { DbStore } from "../db";

export namespace BelongsToDoc {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string, symbol: Symbol) {
        let key = phpDoc.uri + DbStore.uriSep + name;

        return db.put(key, symbol);
    }

    export async function removeByDoc(db: DbStore, uri: string): Promise<string[]> {
        let prefix = uri + DbStore.uriSep;

        return new Promise<string[]>((resolve, reject) => {
            let names: string[] = [];

            db.prefixSearch(prefix)
                .on('data', (data) => {
                    names.push(data.key.substr(data.key.indexOf(DbStore.uriSep)));

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
}