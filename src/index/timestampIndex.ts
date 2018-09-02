import { DbStore } from "../storage/db";
import { inject, named, injectable } from "inversify";
import { IndexId } from "../constant";

@injectable()
export class TimestampIndex {
    private db: DbStore;

    constructor(@inject(BindingIdentifier.DB_STORE) @named(IndexId.TIMESTAMP) store: DbStore) {
        this.db = store;
    }

    async put(uri: string, timestamp: number): Promise<void> {
        return this.db.put(uri, timestamp);
    }

    async get(uri: string): Promise<number | null> {
        let timestamp: number | null = null;

        try {
            timestamp = await this.db.get(uri);
        } catch(err) { }

        return timestamp;
    }

    async delete(uri: string): Promise<void> {
        return this.db.del(uri);
    }
}