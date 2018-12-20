import { PhpDocument } from "../../symbol/phpDocument";
import { Symbol } from "../../symbol/symbol";
import { DbStore } from "../db";

export namespace BelongsToDoc {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string, symbol: Symbol) {
        let key = phpDoc.uri + DbStore.uriSep + name;

        return db.put(key, symbol);
    }

    export async function removeByDoc(db: DbStore, uri: string) {
        let prefix = uri + DbStore.uriSep;

        return new Promise<void>((resolve, reject) => {
            db.prefixSearch(prefix)
                .on('data', (data) => {
                    db.del(data.key);
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve();
                });
        });
    }
}