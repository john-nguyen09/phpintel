import { Location } from "./meta";
import { VisibilityModifier } from "./modifier";

export interface PropertyModifier {
    visibility: VisibilityModifier;
    static: boolean;
}

export class Property {
    public modifier: PropertyModifier = {
        visibility: VisibilityModifier.Public,
        static: false,
    };
    public name: string = '';
    public location: Location | undefined = undefined;
    public type: string = '';
    public scope: string | undefined = undefined;
}