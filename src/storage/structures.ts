export interface DbStoreInfo {
    name: string;
    version: number;
    keyEncoding?: Level.Encoding | string;
    valueEncoding?: Level.Encoding | string;
}