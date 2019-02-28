import { Location } from "./meta";
import { Method } from "./method";
import { Property } from "./property";

export class Trait {
    public name: string = '';
    public location: Location | undefined = undefined;
    public description: string = '';

    public traits: Trait[] = [];

    public methods: Method[] = [];
    public properties: Property[] = [];
}