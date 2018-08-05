import { Range } from "./range";
import { toRelative } from "../../util/uri";

export class Location {
    constructor(public uri: string, public range: Range) { }

    toObject(): any {
        Object.prototype.constructor = this.constructor;
        let object: any = new Object();
        
        object.uri = toRelative(this.uri);
        object.range = this.range;

        return object;
    }
}