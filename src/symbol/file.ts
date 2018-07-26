import { Symbol } from "./symbol";
import { PhpDocument } from "../phpDocument";

export class File extends Symbol {
    public symbols: Symbol[] = [];

    constructor(doc: PhpDocument) {
        super(null, doc);
    }

    consume(other: Symbol) {
        this.symbols.push(other);

        return false;
    }
}