import { Constant } from "./constant";
import { VisibilityModifier } from "./modifier";
import { Type } from "./typeResolver/type";

export class ClassConstant extends Constant {
    public visibility: VisibilityModifier = VisibilityModifier.Public;
    public scope: Type | undefined = undefined;

    public extends(constant: Constant) {
        this.name = constant.name;
        this.value = constant.value;
        this.type = constant.type;
        this.description = constant.description;
        this.location = constant.location;
    }
}