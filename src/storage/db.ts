import level = require("level");
import sublevel = require("subleveldown");
import { DbStoreInfo } from "./structures";
import { injectable, inject } from "inversify";

@injectable()
export class LevelDatasource {
    private db: Level.LevelUp;

    constructor(location: string, options: any) {
        this.db = level(location, options);
    }

    getDb() {
        return this.db;
    }
}

@injectable()
export class DbStore {
    private static readonly separator = '!';
    private static readonly versionPrefix = '@';

    private storeKey: string;
    private db: Level.LevelUp;

    constructor(
        datasource: LevelDatasource,
        storeInfo: DbStoreInfo
    ) {
        this.storeKey = storeInfo.name + DbStore.versionPrefix + storeInfo.version;
        this.db = sublevel(datasource.getDb(), this.storeKey, {
            separator: DbStore.separator,
            valueEncoding: 'json'
        });
    }

    async put(key: string, value: any): Promise<void> {
        return this.db.put(key, value);
    }

    async get(key: string): Promise<any> {
        return this.db.get(key);
    }

    async del(key: string): Promise<void> {
        return this.db.del(key);
    }

    createReadStream(options?: Level.ReadStreamOptions): NodeJS.ReadableStream {
        return this.db.createReadStream(options);
    }

    iterator(options?: Level.ReadStreamOptions): Level.Iterator {
        return this.db.iterator(options);
    }

    prefixSearch(prefix: string): NodeJS.ReadableStream {
        return this.createReadStream({
            gte: prefix,
            lte: prefix + '\xFF'
        });
    }
}