import { DbStore } from "../../db";
import { PhpDocument } from "../../../symbol/phpDocument";

export interface NameIndexData {
    name: string;
    uri: string;
}

export namespace NameIndex {
    export async function put(db: DbStore, phpDoc: PhpDocument, name: string) {
        return db.put(name + DbStore.URI_SEP + phpDoc.uri, phpDoc.uri);
    }

    export async function remove(db: DbStore, uri: string, name: string) {
        return db.del(name + DbStore.URI_SEP + uri);
    }

    export async function get(db: DbStore, name: string): Promise<string[]> {
        return new Promise<string[]>((resolve, reject) => {
            let keys: string[] = [];

            db.prefixSearch(name + DbStore.URI_SEP)
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

    export async function prefixSearch(db: DbStore, prefix: string): Promise<NameIndexData[]> {
        return new Promise<NameIndexData[]>((resolve, reject) => {
            let datas: NameIndexData[] = [];

            db.prefixSearch(prefix)
                .on('data', (data) => {
                    datas.push({
                        name: getNameFromKey(data.key),
                        uri: data.value,
                    });
                })
                .on('error', (err) => {
                    reject(err);
                })
                .on('end', () => {
                    resolve(datas);
                });
        });
    }

    export function getNameFromKey(key: string): string {
        return key.substr(0, key.indexOf(DbStore.URI_SEP));
    }
}