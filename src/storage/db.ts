import level = require("level");
import sublevel = require("subleveldown");
import { DbStoreInfo } from "./structures";
import { injectable, inject } from "inversify";

@injectable()
export class LevelDatasource {
    private db: Level.LevelUp;

    constructor(location: string, options?: any) {
        this.db = level(location, options);
    }

    getDb() {
        return this.db;
    }
}

@injectable()
export class DbStore {
    public static readonly URI_SEP = '#';
    
    protected static readonly separator = '!';
    protected static readonly versionPrefix = '@';

    protected storeKey: string;
    protected db: Level.LevelUp;

    async put(key: string | Buffer, value: any): Promise<void> {
        return this.db.put(key, value);
    }

    async get(key: string | Buffer): Promise<any> {
        return this.db.get(key);
    }

    async del(key: string | Buffer): Promise<void> {
        return this.db.del(key);
    }

    createReadStream(options?: Level.ReadStreamOptions): NodeJS.ReadableStream {
        return this.db.createReadStream(options);
    }

    iterator<T>(options?: Level.ReadStreamOptions): Level.Iterator<T> {
        return this.db.iterator<T>(options);
    }

    prefixSearch(prefix: string, limit?: number): NodeJS.ReadableStream {
        if (typeof limit === 'undefined') {
            limit = -1;
        }

        return this.createReadStream({
            gte: prefix,
            lte: prefix + '\xFF',
            limit: limit
        });
    }
}

@injectable()
export class SubStore extends DbStore {
    constructor(
        datasource: LevelDatasource,
        storeInfo: DbStoreInfo
    ) {
        super();

        this.storeKey = storeInfo.name + DbStore.versionPrefix + storeInfo.version;
        this.db = sublevel(datasource.getDb(), this.storeKey, {
            separator: DbStore.separator,
            keyEncoding: storeInfo.keyEncoding,
            valueEncoding: storeInfo.valueEncoding
        });
    }
}