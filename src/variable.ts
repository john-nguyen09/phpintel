import { Location } from "./meta";

export class Variable {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';
    public type: string = '';
}