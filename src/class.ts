import { Location } from "./meta";
import { Interface } from "./interface";
import { Constant } from "./constant";
import { Property } from "./property";
import { Method } from "./method";
import { ClassModifier } from "./modifier";

export class Class {
    public modifier: ClassModifier = ClassModifier.None;

    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public extends: string[] = [];
    public implements: string[] = [];

    public constants: Constant[] = [];
    public properties: Property[] = [];
    public methods: Method[] = [];
}