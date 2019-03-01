import { Location } from "./meta";
import { Property } from "./property";
import { Method } from "./method";
import { ClassModifier } from "./modifier";
import { ClassConstant } from "./classConstant";

export class Class {
    public modifier: ClassModifier = ClassModifier.None;

    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: string[] = [];
    public implements: string[] = [];

    public constants: ClassConstant[] = [];
    public properties: Property[] = [];
    public methods: Method[] = [];
}