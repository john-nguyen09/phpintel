import { Location } from "./meta";

export class Property {
    public isStatic: boolean = false;
    public name: string = '';
    public location: Location | undefined = undefined;
    public type: string = '';
}