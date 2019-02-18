import { Range } from "./range";

export interface Location {
    uri?: string | undefined;
    range?: Range | undefined;
}
