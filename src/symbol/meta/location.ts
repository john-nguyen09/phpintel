import { Range } from "./range";

export class Location {
    constructor(public uri: string, public range: Range) { }
}