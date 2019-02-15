import { DbStore } from "./db";

export namespace DbHelper {
    export function deleteInStream<V>(
        db: DbStore, stream: NodeJS.ReadableStream, callback?: (data: any) => V
    ): Promise<V[]> {
        const promises: Promise<void>[] = [];
        const results: V[] = [];

        return new Promise<V[]>((resolve, reject) => {
            stream.on('data', data => {
                if (typeof callback !== 'undefined') {
                    results.push(callback(data));
                }

                promises.push(db.del(data.key));
            })
            .on('error', err => {
                Promise.all(promises).then(() => {
                    if (typeof err !== 'undefined') {
                        return reject(err);
                    }

                    return resolve(results);
                })
            })
            .on('end', () => {
                Promise.all(promises).then(() => {
                    resolve(results);
                });
            });
        });
    }
}