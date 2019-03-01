import { VisibilityModifier, StaticModifier, ClassModifier } from "./modifier";
import { Function } from "./function";

export interface MethodModifier {
    visibility: VisibilityModifier;
    static: StaticModifier;
    class: ClassModifier;
}

export class Method extends Function {
    public modifier: MethodModifier = {
        visibility: VisibilityModifier.Public, // Methods have public visibility by default
        static: StaticModifier.None,
        class: ClassModifier.None
    };

    public scope: string | undefined = undefined;

    public extends(theFunction: Function) {
        this.name = theFunction.name;
        this.returnType = theFunction.returnType;
        this.parameters = theFunction.parameters;
    }
}