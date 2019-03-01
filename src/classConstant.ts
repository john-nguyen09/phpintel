import { Constant } from "./constant";
import { VisibilityModifier } from "./modifier";

export class ClassConstant extends Constant {
    public visibility: VisibilityModifier = VisibilityModifier.Public;
    public scope: string | undefined = undefined;

    public extends(constant: Constant) {
        this.name = constant.name;
        this.value = constant.value;
        this.type = constant.type;
        this.description = constant.description;
        this.location = constant.location;
    }
}