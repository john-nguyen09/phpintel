import { Location } from "./meta";
import { Variable } from "./variable";
import { TypeComposite } from "./typeResolver/type";

export class Function {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public parameters: Variable[] = [];

    public returnType: TypeComposite = new TypeComposite();
}