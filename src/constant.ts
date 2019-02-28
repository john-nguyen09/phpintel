import { Location } from "./meta";
import { Interface } from "./interface";
import { Class } from "./class";

export class Constant {
    public name: string = '';
    public value: string = '';
    public type: string = '';
    public location: Location | undefined = undefined;

    public scope: Class | Interface | undefined = undefined;
}