import { Range } from "./range";
import { toRelative } from "../../util/uri";
import { FieldGetter } from "../fieldGetter";
import { nonenumerable } from "../../util/decorator";

export interface Location {
    uri?: string | undefined;
    range?: Range | undefined;
}

// export class Location implements FieldGetter {
//     @nonenumerable
//     public uri: string;
//     public range: Range;

//     constructor(uri?: string, range?: Range) {
//         if (uri !== undefined) {
//             this.uri = uri;
//         }

//         if (range !== undefined) {
//             this.range = range;
//         }
//     }

//     get relativeUri(): string {
//         return toRelative(this.uri);
//     }

//     get isEmpty(): boolean {
//         return typeof this.uri === 'undefined' || typeof this.range === 'undefined';
//     }

//     getFields(): string[] {
//         return [
//             'relativeUri', 'range'
//         ];
//     }
// }