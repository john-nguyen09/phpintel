import { Symbol, Consumer } from "../symbol";
import { NamespaceName } from "./namespaceName";

export class QualifiedName extends Symbol implements Consumer {
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof NamespaceName) {
            this.name = other.name;

            return true;
        }

        return false;
    }
}