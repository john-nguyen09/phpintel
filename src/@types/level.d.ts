// Type definitions for level
// Project: https://github.com/Level/level

declare module 'level' {
    function level(location: string): Level.LevelUp;

    export = level;
}

declare namespace Level {
    export interface LevelUp {
        put(key: string, value: any): Promise<void>;
        get(key: string): Promise<any>;
        del(key: string): Promise<void>;
        batch(ops: BatchOperation[]): Promise<void>;
        createReadStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
        createKeyStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
        createValueStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
    }
    
    export interface BatchOperation {
        type: string;
        key: string;
        value?: any;
    }
    
    export interface ReadStreamOptions {
        gt?: string;
        gte?: string;
        lt?: string;
        lte?: string;
        reverse?: boolean;
        limit?: number;
        keys?: boolean;
        values?: boolean;
    }
}