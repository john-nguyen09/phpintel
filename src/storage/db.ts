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
    public static readonly uriSep = '#';
    
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

@injectable()
export class SubStore extends DbStore {
    constructor(
        datasource: LevelDatasource,
        storeInfo: DbStoreInfo,
        valueEncoding: Level.Encoding
    ) {
        super();

        this.storeKey = storeInfo.name + DbStore.versionPrefix + storeInfo.version;
        this.db = sublevel(datasource.getDb(), this.storeKey, {
            separator: DbStore.separator,
            valueEncoding: valueEncoding
        });
    }
}