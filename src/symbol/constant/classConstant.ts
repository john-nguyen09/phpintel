import { Symbol } from "../symbol";
import { Identifier } from "../identifier";
import { Constant } from "./constant";

export class ClassConstant extends Constant {
    consume(other: Symbol) {
        if (other instanceof Identifier) {
            this.name = other.name;

            return true;
        } else {
            super.consume(other);
        }

        return false;
    }
}