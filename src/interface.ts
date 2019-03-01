import { Location } from "./meta";
import { Method } from "./method";
import { ClassConstant } from "./classConstant";

export class Interface {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: Interface[] = [];

    public constants: ClassConstant[] = [];
    public methods: Method[] = [];
}