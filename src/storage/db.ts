import level = require("level");
import { DbStoreInfo } from "./structures";

export namespace DB {
    var _db: Level.LevelUp;

    export function init(location: string): void {
        _db = level(location);
    }

    export async function put(key: string, value: any): Promise<void> {
        return _db.put(key, JSON.stringify(value));
    }

    export async function get(key: string): Promise<any> {
        let jsonString = await _db.get(key);

        return JSON.parse(jsonString);
    }

    export async function del(key: string): Promise<void> {
        return _db.del(key);
    }

    export function createReadableStream(options: Level.ReadStreamOptions): NodeJS.ReadableStream {
        return _db.createReadStream(options);
    }
}

export class DbStore {
    private static readonly separator = '!';
    private static readonly versionPrefix = '@';

    private storeKey: string;
    private prefix: string;

    constructor(private storeInfo: DbStoreInfo) {
        this.storeKey = storeInfo.name + DbStore.versionPrefix + storeInfo.version;
        this.prefix = this.storeKey + DbStore.separator;
    }

    async put(key: string, value: any): Promise<void> {
        return DB.put(this.prefix + key, value);
    }
    
    async get(key: string, value: any): Promise<any> {
        return DB.get(this.prefix + key);
    }

    async del(key: string): Promise<void> {
        return DB.del(this.prefix + key);
    }

    createReadableStream(options: Level.ReadStreamOptions): NodeJS.ReadableStream {
        const needPrefixList: Extract<keyof Level.ReadStreamOptions, string>[] = [
            'gt', 'gte', 'lt', 'lte'
        ];

        for (let needPrefix of needPrefixList) {
            if (needPrefix in options && options[needPrefix] != undefined) {
                options[needPrefix] = this.prefix + options[needPrefix];
            }
        }

        return DB.createReadableStream(options);
    }

    prefixSearch(prefix: string): NodeJS.ReadableStream {
        return this.createReadableStream({
            gte: prefix,
            lte: prefix + '\xFF'
        });
    }
}