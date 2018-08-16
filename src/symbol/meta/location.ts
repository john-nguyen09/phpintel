import { Range } from "./range";
import { toRelative } from "../../util/uri";
import { FieldGetter } from "../fieldGetter";

export class Location implements FieldGetter {
    constructor(public uri: string, public range: Range) { }

    get relativeUri(): string {
        return toRelative(this.uri);
    }

    getFields(): string[] {
        return [
            'relativeUri', 'range'
        ];
    }
}