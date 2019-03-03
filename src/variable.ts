import { Location } from "./meta";
import { TypeComposite, Type } from "./typeResolver/type";

export class Variable {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';
    public type: TypeComposite = new TypeComposite();
}