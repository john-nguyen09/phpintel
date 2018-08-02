import { Phrase, Parser } from "php7parser";
import { Consumer, Symbol } from "./symbol";

export class PhpDocument extends Symbol implements Consumer {
    public symbols: Symbol[] = [];

    constructor(public uri: string, public text: string) {
        super(null, null);
    }

    getTree(): Phrase {
        return Parser.parse(this.text);
    }

    consume(other: Symbol): boolean {
        this.symbols.push(other);

        return true;
    }
}