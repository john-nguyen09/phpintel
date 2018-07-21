import { Phrase, Parser } from "../node_modules/php7parser";

export class PhpDocument {
    constructor(public uri: string, public text: string) { }

    getTree(): Phrase {
        return Parser.parse(this.text);
    }
}