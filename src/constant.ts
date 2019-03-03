import { Location } from "./meta";
import { Type, TypeComposite } from "./typeResolver/type";

export class Constant {
    public name: Type = new Type('');
    public description: string = '';
    public value: string = '';
    public type: TypeComposite = new TypeComposite();
    public location: Location | undefined = undefined;
}