import { Location } from "./meta";
import { Constant } from "./constant";
import { Method } from "./method";

export class Interface {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: Interface[] = [];

    public constants: Constant[] = [];
    public methods: Method[] = [];
}