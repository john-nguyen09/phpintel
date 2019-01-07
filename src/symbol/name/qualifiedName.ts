import { Symbol, Consumer, Locatable } from "../symbol";
import { NamespaceName } from "./namespaceName";
import { Location } from "../meta/location";

export class QualifiedName extends Symbol implements Consumer, Locatable {
    public name: string = '';
    public location: Location = new Location();

    consume(other: Symbol) {
        if (other instanceof NamespaceName) {
            this.name = other.name;

            return true;
        }

        return false;
    }
}