import { Location } from "./meta";
import { Property } from "./property";
import { Method } from "./method";
import { ClassModifier } from "./modifier";
import { ClassConstant } from "./classConstant";
import { Type } from "./typeResolver/type";

export class Class {
    public modifier: ClassModifier = ClassModifier.None;

    public name: Type = new Type('');
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: Type[] = [];
    public implements: Type[] = [];

    public constants: ClassConstant[] = [];
    public properties: Property[] = [];
    public methods: Method[] = [];
}