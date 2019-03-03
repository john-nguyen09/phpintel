import { Location } from "./meta";
import { Method } from "./method";
import { ClassConstant } from "./classConstant";
import { Type } from "./typeResolver/type";

export class Interface {
    public name: Type | undefined = undefined;
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: Interface[] = [];

    public constants: ClassConstant[] = [];
    public methods: Method[] = [];
}