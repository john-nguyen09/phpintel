import * as crypto from "crypto";
import { injectable } from "inversify";

@injectable()
export class Hasher {
    private hash: crypto.Hash;

    constructor() {
        this.hash = crypto.createHash('md5');
    }

    getHash(str: string) {
        return this.hash.update(str).digest('hex');
    }
}