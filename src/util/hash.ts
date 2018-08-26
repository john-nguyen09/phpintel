import * as crypto from "crypto";

export namespace Hasher {
    var hash: crypto.Hash;

    export function init() {
        hash = crypto.createHash('md5');
    }

    export function getHash(str: string) {
        return hash.update(str).digest('hex');
    }
}