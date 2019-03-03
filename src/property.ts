import { Location } from "./meta";
import { VisibilityModifier } from "./modifier";
import { Type, TypeComposite } from "./typeResolver/type";

export interface PropertyModifier {
    visibility: VisibilityModifier;
    static: boolean;
}

export class Property {
    public modifier: PropertyModifier = {
        visibility: VisibilityModifier.Public,
        static: false,
    };
    public name: Type | undefined = undefined;
    public location: Location | undefined = undefined;
    public type: TypeComposite = new TypeComposite();
    public scope: Type | undefined = undefined;
}