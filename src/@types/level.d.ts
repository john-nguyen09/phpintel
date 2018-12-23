// Type definitions for level
// Project: https://github.com/Level/level

declare module 'level' {
    function level(location: string, options?: any): Level.LevelUp;

    export = level;
}

declare namespace Level {
    export interface LevelUp {
        put(key: string | Buffer, value: any): Promise<void>;
        get(key: string | Buffer): Promise<any>;
        del(key: string | Buffer): Promise<void>;
        batch(ops: BatchOperation[]): Promise<void>;
        createReadStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
        createKeyStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
        createValueStream(options?: ReadStreamOptions): NodeJS.ReadableStream;
        iterator<T>(options?: ReadStreamOptions): Iterator<T>;
    }
    
    export interface BatchOperation {
        type: string;
        key: string | Buffer;
        value?: any;
    }
    
    export interface ReadStreamOptions {
        gt?: string | Buffer;
        gte?: string | Buffer;
        lt?: string | Buffer;
        lte?: string | Buffer;
        reverse?: boolean;
        limit?: number;
        keys?: boolean;
        values?: boolean;
    }

    export type IteratorNextCallback<T> = (
        error: Error | null,
        key: string | Buffer,
        value?: T
    ) => void;

    export interface Iterator<T> {
        next(callback: IteratorNextCallback<T>): void;
        seek(target: any): void;
        end(callback?: () => void): void;
    }

    export interface Encoding {
        type: string;
        encode: (obj: any) => Buffer;
        decode: (buffer: Buffer) => any;
        buffer: boolean;
    }
}