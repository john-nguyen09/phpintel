import { Location } from "./meta";
import { Variable } from "./variable";

export class Function {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public parameters: Variable[] = [];

    public returnType: string = '';
}