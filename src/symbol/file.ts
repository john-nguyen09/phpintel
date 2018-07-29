import { Symbol, Consumer } from "./symbol";
import { PhpDocument } from "../phpDocument";

export class File extends Symbol implements Consumer {
    public symbols: Symbol[] = [];

    constructor(doc: PhpDocument) {
        super(null, doc);
    }

    consume(other: Symbol) {
        this.symbols.push(other);

        return false;
    }
}