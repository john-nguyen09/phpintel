import { Location } from "./meta";
import { Method } from "./method";
import { Property } from "./property";
import { Type } from "./typeResolver/type";

export class Trait {
    public name: Type | undefined = undefined;
    public location: Location | undefined = undefined;
    public description: string = '';

    public traits: Trait[] = [];

    public methods: Method[] = [];
    public properties: Property[] = [];
}